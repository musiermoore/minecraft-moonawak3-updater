# MoonAwak3 Minecraft Updater

Console updater for the MoonAwak3 Minecraft mod pack. The app downloads the public Yandex Disk archive, unpacks it, lets you choose which archive folders/files should be copied into the local `mods` folder, and then replaces the old `mods` folder.

## Requirements

- Go 1.25 or newer
- Internet access for downloading the archive

## Run

```sh
go run .
```

## Build

Build for the current platform:

```sh
go build .
```

Build a Windows executable:

```sh
GOOS=windows GOARCH=amd64 go build -o updater.exe .
```

Build release binaries for Windows, macOS, and Linux:

```sh
sh scripts/build.sh
```

For a new release tag, pass the version into the build:

```sh
VERSION=v0.0.2 sh scripts/build.sh
```

The generated binaries are written to `dist/`.

Upload these files to the matching GitHub Release:

- `moonawak3-minecraft-windows-amd64.exe`
- `moonawak3-minecraft-macos-arm64`
- `moonawak3-minecraft-macos-amd64`
- `moonawak3-minecraft-linux-amd64`
- `checksums.txt`

The app checks GitHub Releases in `musiermoore/minecraft-moonawak3-updater` and compares the latest release tag with its built-in version.

## Selection Controls

The selector shows only the supported top-level archive folders:

- `mods`
- `shaderpacks`
- `Новый мод` / `Новые моды`

Default selected folders:

- `mods`
- `Новый мод` / `Новые моды`

Controls:

- `↑` / `↓`: move the cursor
- `Space`: select or unselect the highlighted item
- `→`: expand or collapse the highlighted top-level folder
- `←`: collapse the expanded top-level folder
- `N`: next page, when pagination is shown
- `P`: previous page, when pagination is shown
- `Enter`: continue with the current selection
- `Esc`: when a folder is open, return to the root view
- `Esc` twice on the root view: close the app

Only top-level folders can be expanded. When a top-level folder is expanded, the selector shows its direct files and folders.

After the updater finishes, it waits for a keypress before closing. If no key is pressed, it closes automatically after 60 seconds.

## Generated Files

The app may create these runtime files and folders:

- `mods.zip`
- `temp_mods/`
- `temp_selected_mods/`
- `mods/`

They are ignored by Git.
