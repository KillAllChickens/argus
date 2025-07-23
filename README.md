<div align="center">

![Banner](assets/ArgusBanner.jpg)

[![License: GPL v3](https://img.shields.io/github/license/KillAllChickens/argus)](https://www.gnu.org/licenses/gpl-3.0)&nbsp;
[![Go Report Card](https://goreportcard.com/badge/github.com/KillAllChickens/argus?cache_bust=1)](https://goreportcard.com/report/github.com/KillAllChickens/argus)&nbsp;
![GitHub stars](https://img.shields.io/github/stars/KillAllChickens/argus)&nbsp;
![GitHub forks](https://img.shields.io/github/forks/KillAllChickens/argus)&nbsp;

</div>

> _"The all-seeing one."_

Named after the hundred-eyed giant of Greek mythology, Argus Panoptes is a powerful OSINT (Open Source Intelligence) tool designed to uncover the digital footprint of a specific username. Just as his mythological namesake served as a vigilant watchman, this tool scans the web to identify websites where a target username is registered.

For better accuracy, Argus Panoptes can use Google Gemini to filter out false positives, making results as precise as possible.

<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->
<details>
<summary>Table of Contents</summary>

  - [✨ Features](#-features)
  - [🛠️ Installation](#-installation)
    - [Linux](#linux)
    - [Windows](#windows)
  - [Usage](#usage)
    - [Configuration](#configuration)
    - [Scanning](#scanning)
  - [📝 Usernames](#-usernames)
    - [Command-Line Usernames](#command-line-usernames)
    - [Username Files](#username-files)
- [🚧 Planned Features](#-planned-features)
  - [Finished](#finished)

</details>
<!-- END doctoc generated TOC please keep comment here to allow auto update -->

## ✨ Features

- 🚀 **Blazing Fast, Multi-threaded Scanning:** In testing, single username scans across **170+ sites** completed in under **5 seconds**.
  - **Note:** Enabling AI-powered scanning will limit the thread count to **5** to prevent rate-limiting, which will result in a significant slowdown.
- 🤖 **AI-Powered False Positive Detection:** Uses Google Gemini for more accurate identification of user profiles.
- 🔧 **Highly Customizable:** Tailor the site list, user agents, soft 404 detection, and even the ASCII art to your preferences.
- 📄 **Flexible Output Formats:** Export scan results in various formats, including PDF, HTML, JSON, and TXT.

## 🛠️ Installation

### Linux

1.  **Install Golang:**
    - **Debian-based (like Ubuntu):**
      ```bash
      sudo add-apt-repository ppa:longsleep/golang-backports
      sudo apt update
      sudo apt install golang-go
      ```
    - **Arch-based:**
      ```bash
      sudo pacman -S go
      ```

2.  **Clone the Repository:**

    ```bash
    git clone https://github.com/KillAllChickens/argus
    cd argus
    ```

3.  **Install Argus:**

    ```bash
    ./scripts/install-linux.sh
    ```

4.  **Get Started:**
    Now you're ready to start using Argus! Check out the [Usage](#usage) section below.

### Windows

1.  **Install Go:**
    - Download and install the latest version of Go for Windows from the [official Go website](https://go.dev/dl/).
    - The installation wizard will handle the setup, including adding Go to your system's PATH.

2.  **Install Git:**
    - Download and install [Git for Windows](https://git-scm.com/download/win). This provides Git Bash, the recommended command line for the following steps.

3.  **Clone the Repository:**
    - Open a new Command Prompt or Git Bash window.
    - Run the following commands:
      ```bash
      git clone https://github.com/KillAllChickens/argus
      cd argus
      ```

4.  **Run the Installer:**

    ```batch
    .\scripts\install-windows.bat
    ```

5.  **Get Started:**
    You're all set! See the [Usage](#usage) section to learn how to run your first scan.

## Usage

### Configuration

To enable the AI-powered false positive detection, you'll need to add your Google Gemini API key.

To configure your API key, simply run:

```bash
argus config
# or for short:
argus c
```

### Scanning

- **Scan for a single user:**

  ```bash
  argus scan <username>
  ```

- **Scan for multiple users:**

  ```bash
  argus scan <user1> <user2> <user3>
  ```

- **Scan usernames from a file:**
  Use a `.txt` file with one username per line. For more details, see the [Usernames](#-usernames) section.

  ```bash
  argus scan -u <filename.txt>
  ```

- **Output to different file types:**

  ```bash
  # Output to HTML (default: results/<username>_results.html)
  argus scan <username> --html

  # Output to PDF (default: results/<username>_results.pdf)
  argus scan <username> --pdf

  # Output to JSON (default: results/<username>_results.json)
  argus scan <username> --json

  # Output to Text (default: results/<username>_results.txt)
  argus scan <username> --txt

  # Output to all supported formats you can use --all
  argus scan <username> --all
  ```

- **Proxy and Tor Support:**
  You can use proxies or Tor for enhanced anonymity! Simply specify a proxy with the `--proxy` flag, or Tor with `--tor`.

  ```bash
  # Use a standard HTTP/SOCKS5 proxy (e.g., http://proxyserver:8888 or socks5://user:pass@proxyserver:port)
  argus scan <username> --proxy "http://1.2.3.4:5678"

  # Route all traffic through the Tor network (requires Tor to be installed and running locally on the default port, 9050)
  argus scan <username> --tor

  # Or use a list of proxies
  argus scan <username> --proxy-list "proxies.txt"

  ```

- **Perform a deep scan:**
  Perform a deep scan to gather more information from found profiles, including descriptions, real names, follow/following counts, and more.
  - **Note:** Current only supports a handful of sites, more WILL be added with newer releases.

  ```bash
  argus scan <username> --deep
  # Or for short
  argus scan <username> -d
  ```

- **Additional Options:**
  For a full list of commands and options, use the help flag:

  ```bash
  argus scan --help
  ```

  ```
  NAME:
     scan - Scan username(s).

  USAGE:
     scan [arguments...]

  OPTIONS:
     --threads int, -t int              Amount of concurrent requests (default: 25)
     --ai                               Use AI to eliminate false positives. (Increases scan time) (default: false)
     --username-list string, -u string  Get usernames to scan, one per line
     --output string, -o string         The directory to output to, defaults to ./results/. if you don't specify a specific type, it will output all types
     --proxy string, -p string          Proxy to use for scanning (e.g., http://proxyserver:8888 or socks5://user:pass@proxyserver:port)
     --proxy-list string, --pl string   List of proxied to use, one per line.
     --tor                              Use Tor for scanning (default: false)
     --silent, -s                       Disable "Scan Complete" notifications. (default: false)
     --deep, -d                         Run a Deep Scan, will try to collect more information (default: false)
     --html                             Output as HTML (default: false)
     --pdf                              Output as PDF (default: false)
     --json                             Output as JSON (default: false)
     --text, --txt                      Output as Text (default: false)
     --all                              Output as all supported types (default: false)
  ```

## 📝 Usernames

### Command-Line Usernames

You can specify usernames directly in the command line after the `scan` command.

Use `{?}` as a wildcard to scan for variations of a username. It will be replaced with `-`, `_`, and nothing.

**Example:**

```bash
# This will scan for "username", "user-name", and "user_name"
argus scan "user{?}name"
```

### Username Files

For bulk scanning, you can provide a text file with one username per line.

- Lines starting with `#` are treated as comments and will be ignored.
- Blank lines are also ignored.

**Example `users.txt`:**

```
# This line will be ignored
user1
user2 # This will also be ignored
user3
```

# 🚧 Planned Features

We're constantly working to make Argus even more powerful and versatile! Here is a peek into what we have planned for future releases!

- **Simple Site List Management:** We want to make it easier to customize the sites Argus scans. This includes commands to **add, remove, and update site configurations** directly, putting more control in your hands.

## Finished

- ~~**Proxy List Support:** Currently, you can use a single proxy. We're planning to add the ability to load a list of proxies and **rotate through them automatically** for enhanced anonymity and resilience against rate limits.~~
- ~~**Deep Scan Mode:** Beyond just finding if a username exists, we aim to add a "deep scan" mode. This feature would attempt to **extract more public information** from found profiles (e.g., linked social media, "about me" sections, public post counts).~~
