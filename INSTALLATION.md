# Installation
You can install this tool either via git or via the pre-built binary.
The git approach is a bit technically involved but definitely recommended.
If you are familiar with a terminal, use git.
If the terminal is new to you, you should still consider using git.
Not only does this allow for faster updates, but this way also allows you to inspect and compile the code yourself to guarantee that this isn't a virus.

When installing using the pre-built binary, you have to trust the distributor that this is a legitimate and unaltered build of the source code.
Annoyingly, this approach requires you to temporarily disable Windows defender, which is why this approach is not recommended.

## Download
### Using git
1. Installing git is required. See the Windows section of the official [installation guide](https://git-scm.com/book/en/v2/Getting-Started-Installing-Git)
2. Clone this repository to a suitable location. See the GitHub [guide for cloning repositories](https://docs.github.com/en/repositories/creating-and-managing-repositories/cloning-a-repository?platform=windows)
3. Install the compiler for the `go` programming language [here](https://go.dev/doc/install)
4. Use the terminal to run `go build` inside the cloned repository
5. (Optional) Move the generated `.exe` file to an accessible location like your desktop

### Using the pre-built binary
1. Disable Windows defender. See the Microsoft guide [here](https://support.microsoft.com/en-us/windows/turn-off-defender-antivirus-protection-in-windows-security-99e6004f-c54c-8509-773c-a4d776b77960)
2. Download the latest release of the tool from the [release page](https://github.com/DEFSAK/chiv-admin-helper/releases) (look under assets)
3. (Optional) Move the downloaded `.exe` file to an accessible location like your desktop
4. Create an exception for this program in Windows defender. See the Microsoft guide [here](https://support.microsoft.com/en-us/windows/add-an-exclusion-to-windows-security-811816c0-4dfd-af4a-47e4-c301afe13b26#:~:text=Go%20to%20Start%20%3E%20Settings%20%3E%20Update,%2C%20file%20types%2C%20or%20process.)
5. ***IMPORTANT*** Enable Windows defender again

## Setup
Once the tool is downloaded, simply double-clicking should open a pop-up terminal window that guides you through setup steps.
Most notably, a credential file is needed to unlock the validation and ban feature.

## Update
When a new version is released, it will show up in the [release page](https://github.com/DEFSAK/chiv-admin-helper/releases).
### Using git
1. Use the terminal to pull the latest version of the code by running `git pull` inside the cloned repository
2. Repeat steps 4 and 5 of the download guide

### using the pre-built binary
1. Repeat all steps of the download guide