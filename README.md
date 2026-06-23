# golnk

Weaponize Windows Shortcut files (.LNK) without a Windows OS

## Install 📡

```bash
go install github.com/edoardottt/golnk@latest
```

## Get Started 🎉

```console
Usage:
 golnk -l shortcut_target [-n description] [-w working_dir] [-a cmd_args] [-i icon_path] -o my_file.lnk [-p]

Options:
 -l, --lnk-target               Specifies the shortcut target
 -o, --output-file              Saves the shortcut to a file
 -n, --name                     Specifies a description for the shortcut
 -w, --working-dir              Specifies the working directory for the command
 -a, --arguments                Specifies the arguments for the command
 -i, --icon                     Specifies the icon path
 -p, --printer-link             Generates a network printer shortcut
```

## Usage 💡

**Create a basic shortcut to an application**.  
This creates a simple shortcut pointing directly to Windows Command Prompt.

```bash
golnk -l "C:\Windows\System32\cmd.exe" -o cmd.lnk
```

**Create a shortcut with arguments, a description and an icon**.  
This passes flags to the executable, adds a hover description and links a custom icon file.

```bash
golnk -l "C:\Program Files\App\app.exe" -a "--verbose --debug" -n "Launch App in Debug Mode" -i "C:\Icons\app.ico" -o debug_app.lnk
```

**Create a shortcut to a network folder share**.  
This points the shortcut to a shared folder hosted on a remote server or IP address.

```bash
golnk -l "\\192.168.1.5\share" -o network_share.lnk
```

**Create a shortcut specifically for a network printer**.  
By adding the `-p` flag, the shortcut is generated using Windows network printer properties.

```bash
golnk -l "\\office-server\Printer-HQ" -p -o office_printer.lnk
```

**Create a shortcut with a custom working (startup) directory**.  
This tells Windows to change to a specific directory (like your project folder) before launching the target command.

```bash
golnk -l "C:\Python3\python.exe" -a "script.py" -w "D:\Projects\PythonSrc" -o run_script.lnk
```

## Changelog 📌

Detailed changes for each release are documented in the [release notes](https://github.com/edoardottt/golnk/releases).

## Contributing 🛠️

Just open an [issue](https://github.com/edoardottt/golnk/issues) / [pull request](https://github.com/edoardottt/golnk/pulls).

Before opening a pull request, download [golangci-lint](https://golangci-lint.run/usage/install/) and run

```bash
golangci-lint run
```

If there aren't errors, go ahead :)

## License 📝

This repository is under [MIT License](https://github.com/edoardottt/golnk/blob/main/LICENSE).  
[edoardottt.com](https://edoardottt.com/) to contact me.
