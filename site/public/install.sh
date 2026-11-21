#!/bin/sh
# blackpie installer: curl -fsSL https://useblackpie.ai/install.sh | sh
#
# macOS: puts blackpie.app in /Applications (~/Applications when that is not
# writable) and links the blackpie command into ~/.local/bin.
# Linux: puts blackpie in ~/.local/bin — the desktop app (with a menu entry)
# when GTK 3 and WebKitGTK 4.1 are installed, the terminal build otherwise.
# A computer with a desktop but without them gets the terminal build and
# the command that installs them on its distribution; --gui and --tui
# choose instead (BLACKPIE_CLI=1 is --tui). Nothing is asked.
# Termux: installs the Android terminal build into $PREFIX/bin.
# Every download is checked against the release's SHA-256.
#
# Options (curl -fsSL https://useblackpie.ai/install.sh | sh -s -- --proxy …):
#   --proxy <url>      a proxy for every download: http://, https://,
#                      socks5:// or socks5h:// (BLACKPIE_PROXY= does the same;
#                      without either, curl's own https_proxy / HTTPS_PROXY /
#                      ALL_PROXY are used as usual)
#   --mirror <prefix>  a GitHub download mirror: the release file is fetched
#                      from <prefix><its github.com URL> (BLACKPIE_MIRROR= does
#                      the same). None is used unless given. The release
#                      list and its SHA-256 still come from useblackpie.ai,
#                      never from the mirror, so a file the mirror changed is
#                      refused.
#   --gui              the desktop app, on Linux even before its libraries are
#                      installed (it opens once they are)
#   --tui              the terminal build (TUI, web UI and CLI), on Linux even
#                      with a desktop (BLACKPIE_CLI=1 does the same)
#   --uninstall        takes off what this script put there: the blackpie
#                      command, and the app (macOS) or its menu entry, icon
#                      and blackpie:// handler (Linux), and the service and
#                      opening at login that would start a blackpie no longer
#                      there. Your settings, providers and usage in
#                      ~/.config/blackpie stay, for when you install again.
#   --purge            --uninstall, and before it `blackpie uninstall --yes`:
#                      every agent disconnected and put back as it was, and
#                      blackpie's settings, providers and usage deleted.
set -eu

site=https://useblackpie.ai
bin="${BLACKPIE_BIN_DIR:-$HOME/.local/bin}"
proxy="${BLACKPIE_PROXY:-}"
mirror="${BLACKPIE_MIRROR:-}"
uninstall=
build=
nolibs=
[ -z "${BLACKPIE_CLI:-}" ] || build=tui

say() { printf '  %s\n' "$*"; }
die() { printf 'blackpie: %s\n' "$*" >&2; exit 1; }

while [ $# -gt 0 ]; do
  case "$1" in
    --proxy) [ $# -ge 2 ] || die "--proxy needs a value"; proxy=$2; shift 2 ;;
    --proxy=*) proxy=${1#--proxy=}; shift ;;
    --mirror) [ $# -ge 2 ] || die "--mirror needs a value"; mirror=$2; shift 2 ;;
    --mirror=*) mirror=${1#--mirror=}; shift ;;
    --uninstall) [ "$uninstall" = purge ] || uninstall=1; shift ;;
    --purge) uninstall=purge; shift ;;
    --gui) build=gui; shift ;;
    --tui|--cli) build=tui; shift ;;
    -h|--help)
      printf '%s\n' "usage: curl -fsSL $site/install.sh | sh -s -- [--gui | --tui] [--proxy <url>] [--mirror <prefix>]" \
        "       curl -fsSL $site/install.sh | sh -s -- --uninstall [--purge]" \
        "  --proxy <url>      http://, https://, socks5:// or socks5h:// proxy for the downloads (or BLACKPIE_PROXY=)" \
        "  --mirror <prefix>  GitHub download mirror put before the release's github.com URL (or BLACKPIE_MIRROR=);" \
        "                     the SHA-256 is still checked against $site's" \
        "  --gui              Linux: the desktop app, even before GTK 3 and WebKitGTK 4.1 are installed" \
        "  --tui              Linux: the terminal build (TUI, web UI, CLI), even on a desktop (or BLACKPIE_CLI=1)" \
        "  --uninstall        take off what this script installed; your settings in ~/.config/blackpie stay" \
        "  --purge            with --uninstall: also disconnect every agent and delete blackpie's settings (blackpie uninstall)"
      exit 0 ;;
    *) die "unknown option $1 (--gui, --tui, --proxy <url>, --mirror <prefix>, --uninstall, --purge)" ;;
  esac
done
case "$proxy" in
  ''|http://?*|https://?*|socks5://?*|socks5h://?*) ;;
  *) die "--proxy is an address like http://127.0.0.1:7890 or socks5://127.0.0.1:1080, not $proxy" ;;
esac
case "$mirror" in
  '') ;;
  http://?*|https://?*) case "$mirror" in */) ;; *) mirror="$mirror/" ;; esac ;;
  *) die "--mirror is an http(s) address put before the github.com URL, like https://mirror.example/, not $mirror" ;;
esac

# get: curl, through the proxy when one was given
get() {
  if [ -n "$proxy" ]; then curl -fsSL --proxy "$proxy" "$@"; else curl -fsSL "$@"; fi
}

case "$(uname -s)" in
  Darwin) os=darwin ;;
  Linux)
    termux="${TERMUX_VERSION:-}"
    case "${PREFIX:-}" in */com.termux/files/usr) termux=1 ;; esac
    if [ -n "$termux" ]; then
      [ -n "${PREFIX:-}" ] || die "Termux requires PREFIX to locate its bin directory"
      os=android
      bin="${BLACKPIE_BIN_DIR:-$PREFIX/bin}"
    else
      os=linux
    fi ;;
  *) die "unsupported system: $(uname -s); see $site" ;;
esac
share="${XDG_DATA_HOME:-$HOME/.local/share}"
config="${XDG_CONFIG_HOME:-$HOME/.config}"

# runs <file> <key>: the program an entry's Exec= or a unit's ExecStart=
# starts, its first word, quoted or not
runs() {
  sed -n "s/^$2=//p" "$1" 2>/dev/null | head -n 1 | sed -e 's/^"\([^"]*\)".*/\1/' -e t -e 's/ .*//'
}

# --uninstall: what the install below puts there, and nothing of the
# user's. The service and login entry go too while they start this
# blackpie: left, they would start a program that is gone at each login.
if [ -n "$uninstall" ]; then
  exe="$bin/blackpie"
  if [ "$uninstall" = purge ]; then
    # blackpie disconnects the agents it connected and deletes its own
    # folders; deleting them here by hand would leave the agents pointed
    # at a gateway that is gone
    [ -x "$exe" ] || die "no blackpie at $exe to disconnect your agents and delete its settings: install it again and run this again, or run --uninstall alone to keep them"
    "$exe" uninstall --yes || die "blackpie uninstall didn't finish (see above), so the blackpie command and app are left: run this again once it does"
  fi
  removed=
  gone() { rm -rf "$1"; removed=1; say "removed $1"; }
  if [ "$os" = darwin ]; then
    for apps in /Applications "$HOME/Applications"; do
      [ ! -d "$apps/blackpie.app" ] || gone "$apps/blackpie.app"
    done
    case "$(readlink "$exe" 2>/dev/null)" in
      */blackpie.app/Contents/MacOS/blackpie) gone "$exe" ;;
    esac
  else
    if [ "$os" = linux ]; then
      unit="$config/systemd/user/blackpie.service"
      if [ "$(runs "$unit" ExecStart)" = "$exe" ]; then
        if ! "$exe" daemon uninstall >/dev/null 2>&1; then
          command -v systemctl >/dev/null && systemctl --user disable --now blackpie.service >/dev/null 2>&1 || true
          rm -f "$unit"
        fi
        removed=1; say "removed blackpie's service ($unit)"
      fi
      login="$config/autostart/blackpie.desktop"
      [ "$(runs "$login" Exec)" != "$exe" ] || gone "$login"
    fi
    [ ! -f "$exe" ] || gone "$exe"
    if [ "$os" = linux ]; then
      apps="$share/applications"
      # the menu entry and its icon while they start this blackpie, or one
      # that is gone; another blackpie's stays
      run=$(runs "$apps/blackpie.desktop" Exec)
      if [ -f "$apps/blackpie.desktop" ] && { [ "$run" = "$exe" ] || [ ! -e "$run" ]; }; then
        gone "$apps/blackpie.desktop"
      fi
      # with no menu entry (this one, or the one blackpie uninstall took),
      # what came with it: the icon, the blackpie:// default xdg-mime set
      # (that line only) and the menu's cache when it lists nothing else
      if [ ! -e "$apps/blackpie.desktop" ]; then
        icons="$share/icons/hicolor/256x256/apps"
        [ ! -f "$icons/blackpie.png" ] || gone "$icons/blackpie.png"
        rmdir "$icons" "$share/icons/hicolor/256x256" "$share/icons/hicolor" "$share/icons" 2>/dev/null || true
        list="$config/mimeapps.list"
        if [ -f "$list" ] && grep -qx 'x-scheme-handler/blackpie=blackpie\.desktop;\{0,1\}' "$list"; then
          grep -vx 'x-scheme-handler/blackpie=blackpie\.desktop;\{0,1\}' "$list" > "$list.blackpie-new" || true
          mv "$list.blackpie-new" "$list"
          # a list left with no entry at all (the one xdg-mime made) goes
          grep -qv -e '^\[.*\]$' -e '^[[:space:]]*$' "$list" || rm -f "$list"
          removed=1; say "removed blackpie:// from $list"
        fi
        if [ -d "$apps" ] && ! ls "$apps" | grep -q '\.desktop$'; then
          rm -f "$apps/mimeinfo.cache"
          rmdir "$apps" 2>/dev/null || true
        elif [ -d "$apps" ]; then
          command -v update-desktop-database >/dev/null && update-desktop-database "$apps" 2>/dev/null || true
        fi
      fi
    fi
  fi
  [ "$removed" ] || say "no blackpie installed by this script at $exe"
  if [ "$uninstall" != purge ]; then
    say "kept your settings, providers and usage in $config/blackpie: install again to have them back, or run with --purge to delete them"
    if pid=$(pgrep -f "^$exe( |\$)" 2>/dev/null) && [ -n "$pid" ]; then
      say "blackpie is still running (pid $(printf '%s' "$pid" | tr '\n' ' ' | sed 's/ $//')) and ends when you quit it; agents connected to it reach its gateway until then"
    fi
  fi
  # blackpie uninstall said so already
  [ "$uninstall" = purge ] || say "blackpie is uninstalled"
  exit 0
fi

# has_gui_libs: GTK 3 and WebKitGTK 4.1 (which brings libsoup 3 and
# JavaScriptCore) can be loaded, which is all the desktop app links beyond
# glibc. ldconfig is in /sbin, off a Debian user's PATH, so it is looked for
# there too; with no cache to ask (NixOS, Guix), the usual folders are.
# musl: the desktop app is built against glibc and can't start on musl
# (Alpine); the terminal build is static and runs anywhere
musl() {
  for l in /lib/ld-musl-*; do [ ! -e "$l" ] || return 0; done
  return 1
}
has_gui_libs() {
  libs=
  for l in "$(command -v ldconfig 2>/dev/null)" /sbin/ldconfig /usr/sbin/ldconfig; do
    [ -n "$l" ] && [ -x "$l" ] || continue
    libs=$("$l" -p 2>/dev/null) && [ -n "$libs" ] && break
    libs=
  done
  if [ -n "$libs" ]; then
    printf '%s\n' "$libs" | grep -q 'libwebkit2gtk-4\.1\.so\.0 ' &&
      printf '%s\n' "$libs" | grep -q 'libgtk-3\.so\.0 '
    return
  fi
  for want in libwebkit2gtk-4.1.so.0 libgtk-3.so.0; do
    found=
    for d in /usr/lib/*-linux-gnu /lib/*-linux-gnu /usr/lib64 /lib64 /usr/lib /lib; do
      [ ! -e "$d/$want" ] || { found=1; break; }
    done
    [ -n "$found" ] || return 1
  done
}

# has_desktop: a graphical session here (DISPLAY, WAYLAND_DISPLAY, a
# desktop's variables), a graphical login of this user that an ssh shell
# doesn't inherit (logind), or a desktop installed to log in to
has_desktop() {
  [ -z "${DISPLAY:-}${WAYLAND_DISPLAY:-}${XDG_CURRENT_DESKTOP:-}${XDG_SESSION_DESKTOP:-}${DESKTOP_SESSION:-}" ] || return 0
  case "${XDG_SESSION_TYPE:-}" in x11|wayland|mir) return 0 ;; esac
  if command -v loginctl >/dev/null 2>&1; then
    # Display falls back to a tty login when there is no graphical one,
    # so the session's own type is what tells
    s=$(loginctl show-user "$(id -un 2>/dev/null)" -p Display --value 2>/dev/null) || s=
    if [ -n "$s" ]; then
      case "$(loginctl show-session "$s" -p Type --value 2>/dev/null)" in
        x11|wayland|mir) return 0 ;;
      esac
    fi
  fi
  dirs=${XDG_DATA_DIRS:-/usr/local/share:/usr/share}
  old=$IFS; IFS=:
  for d in $dirs; do
    for f in "$d"/xsessions/*.desktop "$d"/wayland-sessions/*.desktop; do
      [ ! -e "$f" ] || { IFS=$old; return 0; }
    done
  done
  IFS=$old
  return 1
}

# gui_packages: the command that installs the desktop app's libraries on
# this distribution, or nothing when it isn't one we know
gui_packages() {
  id='' like=''
  # BLACKPIE_OS_RELEASE: another os-release, for the installer's tests
  rel=${BLACKPIE_OS_RELEASE:-/etc/os-release}
  [ -r "$rel" ] || rel=/usr/lib/os-release
  if [ -r "$rel" ]; then
    id=$(sed -n 's/^ID=//p' "$rel" | tr -d "\"'" | head -n 1)
    like=$(sed -n 's/^ID_LIKE=//p' "$rel" | tr -d "\"'" | head -n 1)
  fi
  for d in $id $like; do
    case "$d" in
      debian|ubuntu|linuxmint|pop|elementary|zorin|kali|raspbian|deepin|uos)
        echo "sudo apt install libwebkit2gtk-4.1-0"; return ;;
      fedora|rhel|centos|rocky|almalinux|nobara|openeuler|anolis)
        echo "sudo dnf install webkit2gtk4.1"; return ;;
      arch|manjaro|endeavouros|cachyos|garuda)
        echo "sudo pacman -S --needed webkit2gtk-4.1"; return ;;
      opensuse*|suse|sles)
        echo "sudo zypper install libwebkit2gtk-4_1-0"; return ;;
      void)
        echo "sudo xbps-install -S libwebkit2gtk41"; return ;;
      gentoo)
        echo "sudo emerge net-libs/webkit-gtk:4.1"; return ;;
    esac
  done
}

case "$(uname -m)" in
  arm64|aarch64) arch=arm64 ;;
  x86_64|amd64) arch=amd64 ;;
  *) die "unsupported CPU: $(uname -m)" ;;
esac
case "$os" in
  darwin) file="blackpie-darwin-$arch.zip" ;;
  android) file="blackpie-cli-android-$arch" ;;
  *)
    file="blackpie-cli-linux-$arch"
    case "$build" in
      gui) ! musl || die "the desktop app is built for glibc and can't run on musl (Alpine): install the terminal build with --tui"
           file="blackpie-linux-$arch" ;;
      tui) ;;
      *) if musl; then :
         elif has_gui_libs; then file="blackpie-linux-$arch"
         elif has_desktop; then nolibs=1; fi ;;
    esac ;;
esac

tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT

# the release list and its checksums: always from the site, never a mirror
feed=$(get "$site/api/latest") || die "could not reach $site${proxy:+ through $proxy}"
# The feed is one line of JSON; pull this file's url and hash out of it.
entry=$(printf '%s' "$feed" | sed -n "s/.*\"$file\":{\([^}]*\)}.*/\1/p")
url=$(printf '%s' "$entry" | sed -n 's/.*"url":"\([^"]*\)".*/\1/p')
sum=$(printf '%s' "$entry" | sed -n 's/.*"sha256":"\([^"]*\)".*/\1/p')
version=$(printf '%s' "$feed" | sed -n 's/.*"version":"\([^"]*\)".*/\1/p')
if [ "$os" = android ] && { [ -z "$url" ] || [ -z "$sum" ]; }; then
  die "the newest release has no $file; build in Termux with go build -tags nogui (see $site/docs/start)"
fi
[ -n "$url" ] && [ -n "$sum" ] || die "the newest release has no $file"

from=$url
case "$url" in
  https://github.com/*) [ -z "$mirror" ] || from="$mirror$url" ;;
esac
if [ "$from" != "$url" ]; then say "downloading blackpie $version ($file) through $mirror"
else say "downloading blackpie $version ($file)"; fi
get -o "$tmp/$file" "$from" || die "could not download $from"
if command -v sha256sum >/dev/null; then got=$(sha256sum "$tmp/$file" | cut -d' ' -f1)
else got=$(shasum -a 256 "$tmp/$file" | cut -d' ' -f1); fi
if [ "$got" != "$sum" ]; then
  [ "$from" = "$url" ] || die "$file from $mirror does not match its checksum from $site; not installed"
  die "$file does not match its checksum"
fi

mkdir -p "$bin"
if [ "$os" = darwin ]; then
  apps=/Applications
  [ -w "$apps" ] || { apps="$HOME/Applications"; mkdir -p "$apps"; }
  ditto -x -k "$tmp/$file" "$tmp/x"
  # the new app beside the old one first: a move failing (a full disk)
  # leaves the old app in place
  rm -rf "$apps/.blackpie.app.new"
  mv "$tmp/x/blackpie.app" "$apps/.blackpie.app.new"
  rm -rf "$apps/blackpie.app"
  mv "$apps/.blackpie.app.new" "$apps/blackpie.app"
  ln -sf "$apps/blackpie.app/Contents/MacOS/blackpie" "$bin/blackpie"
  say "installed $apps/blackpie.app"
else
  install -m 755 "$tmp/$file" "$bin/blackpie"
  if [ "$file" = "blackpie-linux-$arch" ]; then
    mkdir -p "$share/applications" "$share/icons/hicolor/256x256/apps"
    get -o "$share/icons/hicolor/256x256/apps/blackpie.png" "$site/img/icon-256.png" || true
    cat > "$share/applications/blackpie.desktop" <<EOF
[Desktop Entry]
Type=Application
Name=blackpie
Comment=Every agent's model. One place.
Exec=$bin/blackpie %u
Icon=blackpie
StartupWMClass=blackpie
Categories=Development;Utility;
MimeType=x-scheme-handler/blackpie;
Terminal=false
EOF
    # blackpie://import links open in blackpie
    command -v update-desktop-database >/dev/null && update-desktop-database "$share/applications" 2>/dev/null || true
    command -v xdg-mime >/dev/null && xdg-mime default blackpie.desktop x-scheme-handler/blackpie 2>/dev/null || true
    say "installed the blackpie desktop app (menu entry: blackpie)"
  elif [ "$os" = android ]; then
    say "installed the Termux build; use blackpie web for the browser UI or blackpie tui for the terminal UI"
  else
    say "installed the terminal build (TUI, web UI and CLI)"
  fi
fi
say "installed $bin/blackpie"
case ":$PATH:" in
  *":$bin:"*) ;;
  *) say "add $bin to your PATH to run blackpie from a terminal" ;;
esac
case "$file" in
  *.zip) say "open it: open -a blackpie" ;;
  blackpie-cli-android-*) say "run it: blackpie web" ;;
  blackpie-linux-*) say "open it from your app menu, or run: blackpie app" ;;
  *) say "run it: blackpie" ;;
esac
# the desktop app without what it needs: what installs it, said last so it
# is what the user reads
if [ "$file" = "blackpie-linux-$arch" ] && ! has_gui_libs; then
  pkgs=$(gui_packages)
  say ""
  say "the desktop app needs GTK 3 and WebKitGTK 4.1, which aren't installed; it opens once they are:"
  say "  ${pkgs:-install WebKitGTK 4.1 (webkit2gtk 4.1) from your distribution}"
elif [ -n "${nolibs:-}" ]; then
  pkgs=$(gui_packages)
  say ""
  say "this computer has a desktop, but the desktop app needs GTK 3 and WebKitGTK 4.1, which aren't installed,"
  say "so the terminal build was installed. For the desktop app, install them and run this again:"
  say "  ${pkgs:-install WebKitGTK 4.1 (webkit2gtk 4.1) from your distribution}"
  say "  curl -fsSL $site/install.sh | sh"
  say "or install the desktop app now and the libraries after: curl -fsSL $site/install.sh | sh -s -- --gui"
fi
