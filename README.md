# [openrepl](https://repl.techmeowt.com)[![Gitter](https://badges.gitter.im/openrepl/openrepl.svg)](https://gitter.im/openrepl/openrepl?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge)
OpenRepl REPL is an online REPL where you can run code in a variety of programming languages on the web.

## How to use
Towards the write is an editor. There, you can write code.
Press run to run this code, and a terminal will pop up below the editor running your code.
Press "Switch Language" to change to a different programming language.

There is also an interactive terminal in the selected programming language on the right.

## Deploying
To deploy a copy of the site, you will need a working Docker install with Docker Compose.
Run the following command in the repo root to deploy:
```
make && docker-compose up -d
```
To update, simply rerun this command in the updated repo.

To tear down the deployment, you can run this command in the repo root:
```
docker-compose down
```

## Editor keybinding
* Ctrl/Cmd-S - save
* Ctrl/Cmd-R - run
* Ctrl/Cmd-F - find

## Adding examples
To add an example, add a source code file to `server/examples/examples` or a subdirectory of this.
To tag this example, add another file with the same name and the secondary extension `.tags`.
Each line in a `.tags` file will be interpreted as a seperate tag.
Blank lines in a `.tags` file will be ignored.
After the examples have been added, redeploy the REPL.
