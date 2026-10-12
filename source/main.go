// Dashboard: one page to switch between the Teams projects and run them.
//
// Every project is a folder in projects\ with a project.json that says which
// program to run and which buttons to show. Projects come prebuilt, so there
// is nothing to compile or move. Several projects can run at the same time;
// click one in the sidebar to see its live log and buttons.
package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

type action struct {
	Label   string   `json:"label"`
	Args    []string `json:"args"`
	Primary bool     `json:"primary,omitempty"` // highlighted button
	Note    string   `json:"note,omitempty"`
}

type project struct {
	ID         string   `json:"-"`
	Name       string   `json:"name"`
	Group      string   `json:"group"`
	About      string   `json:"about"`
	Exe        string   `json:"exe"`                  // relative to the project folder (or dir)
	Dir        string   `json:"dir,omitempty"`        // optional working folder (absolute, or relative to the project folder)
	Screenshot string   `json:"screenshot,omitempty"` // relative to the working folder
	Actions    []action `json:"actions"`
	Order      int      `json:"order"`
	folder     string
}

type run struct {
	Action   string    `json:"action"`
	Started  time.Time `json:"started"`
	Ended    time.Time `json:"ended"`
	Running  bool      `json:"running"`
	Waiting  bool      `json:"waiting"` // program says "Press Enter"
	Exit     int       `json:"exit"`
	Stopped  bool      `json:"stopped"`
	Log      []string  `json:"-"`
	cmd      *exec.Cmd
	stdin    io.WriteCloser
	finished bool
}

var (
	exeDir, projDir string
	mu              sync.Mutex
	projects        []project
	runs            = map[string]*run{}
	waitRe          = regexp.MustCompile(`(?i)press enter`)
)

// ---------- projects ----------

func loadProjects() ([]project, []string) {
	var out []project
	var problems []string
	ents, _ := os.ReadDir(projDir)
	for _, e := range ents {
		if !e.IsDir() {
			continue
		}
		folder := filepath.Join(projDir, e.Name())
		b, err := os.ReadFile(filepath.Join(folder, "project.json"))
		if err != nil {
			continue // not a project (e.g. managed, runs, artifacts)
		}
		var p project
		if err := json.Unmarshal(b, &p); err != nil {
			problems = append(problems, e.Name()+": project.json is not valid JSON ("+err.Error()+")")
			continue
		}
		p.ID, p.folder = e.Name(), folder
		if p.Name == "" {
			p.Name = e.Name()
		}
		out = append(out, p)
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Order != out[j].Order {
			return out[i].Order < out[j].Order
		}
		return out[i].Name < out[j].Name
	})
	return out, problems
}

func (p project) workDir() string {
	if p.Dir == "" {
		return p.folder
	}
	if filepath.IsAbs(p.Dir) {
		return p.Dir
	}
	return filepath.Join(p.folder, p.Dir)
}

func (p project) exePath() string {
	if filepath.IsAbs(p.Exe) {
		return p.Exe
	}
	for _, base := range []string{p.folder, p.workDir()} {
		c := filepath.Join(base, p.Exe)
		if runtime.GOOS != "windows" {
			if t := strings.TrimSuffix(c, ".exe"); fileOK(t) {
				return t
			}
		}
		if fileOK(c) {
			return c
		}
	}
	return filepath.Join(p.folder, p.Exe)
}

func find(id string) (project, bool) {
	mu.Lock()
	defer mu.Unlock()
	for _, p := range projects {
		if p.ID == id {
			return p, true
		}
	}
	return project{}, false
}

func fileOK(p string) bool { st, err := os.Stat(p); return err == nil && !st.IsDir() }

func engineInstalled() bool {
	ents, err := os.ReadDir(filepath.Join(projDir, "managed", "playwright"))
	return err == nil && len(ents) > 0
}

// ---------- running ----------

func (r *run) add(s string) {
	line := time.Now().Format("15:04:05") + "  " + s
	r.Log = append(r.Log, line)
	if len(r.Log) > 4000 {
		r.Log = r.Log[len(r.Log)-4000:]
	}
	if waitRe.MatchString(s) {
		r.Waiting = true
	}
}

func start(p project, a action) error {
	mu.Lock()
	if r := runs[p.ID]; r != nil && r.Running {
		mu.Unlock()
		return errors.New(p.Name + " is already running. Press Stop or Finish first.")
	}
	if isInstall(a) {
		for id, r := range runs {
			if r.Running {
				mu.Unlock()
				return fmt.Errorf("stop the running project (%s) before installing the engine", id)
			}
		}
	} else {
		for id, r := range runs {
			if r.Running && isInstall(action{Args: argsOf(r.Action)}) {
				mu.Unlock()
				return fmt.Errorf("wait for the engine install (%s) to finish", id)
			}
		}
	}
	exe := p.exePath()
	if !fileOK(exe) {
		mu.Unlock()
		return fmt.Errorf("can't find %s", exe)
	}
	c := exec.Command(exe, a.Args...)
	c.Dir = p.workDir()
	pr, pw := io.Pipe()
	c.Stdout, c.Stderr = pw, pw
	in, _ := c.StdinPipe()
	r := &run{Action: a.Label + "\x00" + strings.Join(a.Args, "\x01"), Started: time.Now(), Running: true, cmd: c, stdin: in, Exit: -1}
	r.add(fmt.Sprintf("▶ %s   (%s %s)", a.Label, filepath.Base(exe), strings.Join(a.Args, " ")))
	runs[p.ID] = r
	mu.Unlock()

	if err := c.Start(); err != nil {
		mu.Lock()
		r.add("could not start: " + err.Error())
		r.Running, r.Exit, r.Ended = false, 127, time.Now()
		mu.Unlock()
		return nil
	}
	go func() {
		sc := bufio.NewScanner(pr)
		sc.Buffer(make([]byte, 64*1024), 1024*1024)
		for sc.Scan() {
			mu.Lock()
			r.add(sc.Text())
			mu.Unlock()
		}
	}()
	go func() {
		err := c.Wait()
		pw.Close()
		time.Sleep(150 * time.Millisecond) // let the last lines arrive
		code := 0
		var ee *exec.ExitError
		if errors.As(err, &ee) {
			code = ee.ExitCode()
		} else if err != nil {
			code = 1
		}
		mu.Lock()
		r.Running, r.Waiting, r.Exit, r.Ended = false, false, code, time.Now()
		switch {
		case r.Stopped:
			r.add("■ stopped by you")
		case code == 0:
			r.add("✔ finished OK")
		default:
			r.add(fmt.Sprintf("✖ ended with exit code %d", code))
		}
		logCopy := strings.Join(r.Log, "\n")
		mu.Unlock()
		_ = os.WriteFile(filepath.Join(p.folder, "last-run.log"), []byte(logCopy+"\n"), 0o644)
	}()
	return nil
}

func argsOf(packed string) []string {
	i := strings.IndexByte(packed, 0)
	if i < 0 || i == len(packed)-1 {
		return nil
	}
	return strings.Split(packed[i+1:], "\x01")
}
func labelOf(packed string) string {
	if i := strings.IndexByte(packed, 0); i >= 0 {
		return packed[:i]
	}
	return packed
}
func isInstall(a action) bool {
	for _, x := range a.Args {
		if x == "-install" || x == "-reinstall" {
			return true
		}
	}
	return false
}

func stop(id string) {
	mu.Lock()
	r := runs[id]
	if r == nil || !r.Running || r.cmd == nil || r.cmd.Process == nil {
		mu.Unlock()
		return
	}
	r.Stopped = true
	pid := r.cmd.Process.Pid
	r.add("Stopping (closes the program and its browser)...")
	mu.Unlock()
	if runtime.GOOS == "windows" {
		_ = exec.Command("taskkill", "/T", "/F", "/PID", strconv.Itoa(pid)).Run()
	} else {
		_ = r.cmd.Process.Kill()
	}
}

func finish(id string) {
	mu.Lock()
	r := runs[id]
	if r != nil && r.Running && r.stdin != nil {
		r.add("Finish: pressing Enter for you")
		_, _ = r.stdin.Write([]byte("\r\n"))
	}
	mu.Unlock()
}

// ---------- web ----------

type projView struct {
	ID, Name, Group, About string
	Actions                []action
	State                  string // idle / running / waiting / ok / failed / stopped
	Current                string
	Since                  string
	HasShot                bool
	Exe                    string
	Folder                 string
	ExeFound               bool
}

func state(r *run) (string, string, string) {
	if r == nil {
		return "idle", "", ""
	}
	lbl := labelOf(r.Action)
	switch {
	case r.Running && r.Waiting:
		return "waiting", lbl, r.Started.Format("3:04:05 PM")
	case r.Running:
		return "running", lbl, r.Started.Format("3:04:05 PM")
	case r.Stopped:
		return "stopped", lbl, r.Ended.Format("3:04:05 PM")
	case r.Exit == 0:
		return "ok", lbl, r.Ended.Format("3:04:05 PM")
	}
	return "failed", lbl, r.Ended.Format("3:04:05 PM")
}

func jsonOut(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}

func main() {
	exe, err := os.Executable()
	if err != nil {
		log.Fatal(err)
	}
	exeDir = filepath.Dir(exe)
	projDir = filepath.Join(exeDir, "projects")
	_ = os.MkdirAll(projDir, 0o755)
	var problems []string
	projects, problems = loadProjects()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		io.WriteString(w, page)
	})

	http.HandleFunc("/api/state", func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		var vs []projView
		for _, p := range projects {
			st, cur, since := state(runs[p.ID])
			shot := p.Screenshot != "" && fileOK(filepath.Join(p.workDir(), p.Screenshot))
			vs = append(vs, projView{p.ID, p.Name, p.Group, p.About, p.Actions, st, cur, since, shot,
				p.Exe, p.workDir(), fileOK(p.exePath())})
		}
		pr := append([]string(nil), problems...)
		mu.Unlock()
		jsonOut(w, map[string]any{"projects": vs, "engine": engineInstalled(), "folder": projDir, "problems": pr})
	})

	http.HandleFunc("/api/log", func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("p")
		from, _ := strconv.Atoi(r.URL.Query().Get("from"))
		mu.Lock()
		var lines []string
		n := 0
		if ru := runs[id]; ru != nil {
			n = len(ru.Log)
			if from < 0 || from > n {
				from = 0
			}
			lines = append(lines, ru.Log[from:]...)
		}
		mu.Unlock()
		jsonOut(w, map[string]any{"lines": lines, "next": n})
	})

	http.HandleFunc("/api/shot", func(w http.ResponseWriter, r *http.Request) {
		p, ok := find(r.URL.Query().Get("p"))
		if !ok || p.Screenshot == "" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Cache-Control", "no-store")
		http.ServeFile(w, r, filepath.Join(p.workDir(), p.Screenshot))
	})

	post := func(path string, h func(w http.ResponseWriter, r *http.Request)) {
		http.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				http.Error(w, "POST only", 405)
				return
			}
			if o := r.Header.Get("Origin"); o != "" && o != "http://"+r.Host {
				http.Error(w, "bad origin", 403)
				return
			}
			r.Body = http.MaxBytesReader(w, r.Body, 64*1024)
			h(w, r)
		})
	}

	post("/api/run", func(w http.ResponseWriter, r *http.Request) {
		p, ok := find(r.URL.Query().Get("p"))
		i, _ := strconv.Atoi(r.URL.Query().Get("a"))
		if !ok || i < 0 || i >= len(p.Actions) {
			http.Error(w, "unknown project or button", 400)
			return
		}
		if err := start(p, p.Actions[i]); err != nil {
			http.Error(w, err.Error(), 409)
		}
	})
	post("/api/stop", func(w http.ResponseWriter, r *http.Request) { stop(r.URL.Query().Get("p")) })
	post("/api/finish", func(w http.ResponseWriter, r *http.Request) { finish(r.URL.Query().Get("p")) })
	post("/api/stopall", func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		var ids []string
		for id, ru := range runs {
			if ru.Running {
				ids = append(ids, id)
			}
		}
		mu.Unlock()
		for _, id := range ids {
			stop(id)
		}
	})
	post("/api/rescan", func(w http.ResponseWriter, r *http.Request) {
		ps, pr := loadProjects()
		mu.Lock()
		projects, problems = ps, pr
		mu.Unlock()
	})
	post("/api/install", func(w http.ResponseWriter, r *http.Request) {
		// any project that has an install button can install the shared engine
		mu.Lock()
		ps := append([]project(nil), projects...)
		mu.Unlock()
		for _, p := range ps {
			for _, a := range p.Actions {
				if isInstall(a) {
					if err := start(p, a); err != nil {
						http.Error(w, err.Error(), 409)
					} else {
						jsonOut(w, map[string]string{"project": p.ID})
					}
					return
				}
			}
		}
		http.Error(w, "no project has an install button", 400)
	})
	post("/api/add", func(w http.ResponseWriter, r *http.Request) {
		var in struct{ Name, Folder, Exe, Args, Label string }
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			http.Error(w, "bad form", 400)
			return
		}
		in.Folder = strings.Trim(strings.TrimSpace(in.Folder), `"`)
		in.Exe = strings.Trim(strings.TrimSpace(in.Exe), `"`)
		if in.Name == "" || in.Folder == "" || in.Exe == "" {
			http.Error(w, "name, folder and program are all needed", 400)
			return
		}
		if !filepath.IsAbs(in.Folder) {
			http.Error(w, "folder must be a full path, e.g. C:\\Users\\you\\Teams-Browser-Test\\prototypes\\round2\\playwright", 400)
			return
		}
		if st, err := os.Stat(in.Folder); err != nil || !st.IsDir() {
			http.Error(w, "that folder doesn't exist: "+in.Folder, 400)
			return
		}
		exePath := in.Exe
		if !filepath.IsAbs(exePath) {
			exePath = filepath.Join(in.Folder, in.Exe)
		}
		if !fileOK(exePath) {
			http.Error(w, "can't find the program: "+exePath, 400)
			return
		}
		lbl := in.Label
		if lbl == "" {
			lbl = "Run"
		}
		p := project{Name: in.Name, Group: "Added by you", About: "Runs " + exePath, Exe: exePath, Dir: in.Folder,
			Order: 1000, Actions: []action{{Label: lbl, Args: strings.Fields(in.Args), Primary: true}}}
		slug := strings.Trim(regexp.MustCompile(`[^a-z0-9]+`).ReplaceAllString(strings.ToLower(in.Name), "-"), "-")
		if slug == "" {
			slug = "project"
		}
		dir := filepath.Join(projDir, "added-"+slug)
		if _, err := os.Stat(dir); err == nil {
			http.Error(w, "a project with that name already exists", 409)
			return
		}
		_ = os.MkdirAll(dir, 0o755)
		b, _ := json.MarshalIndent(p, "", "  ")
		if err := os.WriteFile(filepath.Join(dir, "project.json"), b, 0o644); err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		ps, pr := loadProjects()
		mu.Lock()
		projects, problems = ps, pr
		mu.Unlock()
		jsonOut(w, map[string]string{"id": "added-" + slug})
	})
	post("/api/open", func(w http.ResponseWriter, r *http.Request) {
		target := projDir
		if p, ok := find(r.URL.Query().Get("p")); ok {
			target = p.workDir()
		}
		openPath(target)
	})

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		log.Fatal(err)
	}
	url := "http://" + ln.Addr().String() + "/"
	fmt.Println("Dashboard is open in your browser:", url)
	fmt.Println("Keep this window open. Closing it stops every running project.")
	fmt.Printf("Projects folder: %s (%d projects)\n", projDir, len(projects))
	for _, p := range problems {
		fmt.Println("  problem:", p)
	}
	openPath(url)
	log.Fatal(http.Serve(ln, nil))
}

func openPath(target string) {
	var c *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		if strings.HasPrefix(target, "http") {
			c = exec.Command("rundll32", "url.dll,FileProtocolHandler", target)
		} else {
			c = exec.Command("explorer", target)
		}
	case "darwin":
		c = exec.Command("open", target)
	default:
		c = exec.Command("xdg-open", target)
	}
	_ = c.Start()
}
