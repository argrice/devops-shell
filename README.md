# devops-shell

devops-shell is a lightweight shell written in go.lang with built-ins specifically catering towards Dev-Ops Engineers. 

## Features

- Feature 1: Built-in 'runparallel' which can execute multiple commands at once and report live statuses of each job as either 'success', 'running', or 'failed' with a summary of output and status at the end.
- Feature 2: Written in go.lang for use of multi-threading
- Feature 3: Built-in bash shell lookups for autocompletion
- Feature 4: In-memory history as well as permanent input history in home directory

## Missing Features

- Feature 1: Support for | and & operators
- Feature 2: Exiting stuck processes with 'Ctrl+C'
- Feature 3: Built-in 'runremote' for running processes on different hosts
- Feature 4: Built-in 'runremoteparallel' for running multiple processes on different hosts at once
  
### Prerequisites

- Go.lang

### Installation Steps

1. Clone the repository:
   ```bash
   git clone https://github.com/argric/devops-shell.git
   
2. Build the executable:
   ```bash
   cd ~/devops-shell
   go build -o devops-shell

3. Run executable
   ```bash
   ./devops-shell
