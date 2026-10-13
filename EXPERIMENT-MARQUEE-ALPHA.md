# Marquee Alpha integration experiment — local build receipt

This branch was created **from Dashboard `alpha`**, leaving `alpha` unmodified.

The full Marquee integration is provided as the **Dashboard-Marquee-Alpha-Experiment.zip** attachment in the associated ChatGPT conversation. The source package is not yet committed to this GitHub branch; a checkout of this branch alone does **not** run the Marquee experiment.

## What was tested

- Go unit/build package validation on Linux (Go 1.23.2), no test files present
- Windows cross-compile completed (`GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build .`)
- Local HTTP smoke test: two fixed project cards, Dashboard and Teams, both targeting `alpha`
- POST to add arbitrary projects rejected with HTTP 403

## Run supplied package on Windows

Extract **Dashboard-Marquee-Alpha-Experiment.zip** to `C:\Dev\Projects`. Enter the extracted folder and double-click `Start-Marquee-Alpha.cmd`. Visit http://127.0.0.1:17777/.

On selecting Start, each project clones its current GitHub `alpha` into a new experiment-managed build directory. Dashboard listens on 18877; Teams on 18998. The existing approved Alpha remains untouched.

Teams is a private GitHub repository; configure existing Git authentication separately before clicking Start. The experiment refuses automatic Go downloads; missing dependencies are surfaced as errors.

## Next gate

Owner Windows Start / live app / Stop test, then commit audited integration source to this branch and produce a draft comparison against Alpha **before** any promotion. Do not promote based on cross-compilation or HTTP fixture tests.

Safety: no `git clean`, forced checkout, automated upgrade, production profile reuse, or arbitrary project definitions in the experimental build. 
