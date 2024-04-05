#!/bin/bash

# github repo
REPO=RazrFalcon/resvg

SRC=$(realpath $(cd -P "$(dirname "${BASH_SOURCE[0]}")" && pwd))

set -e

WORKDIR=$HOME/src/charts

BUILD_TARGETS=""

OPTIND=1
while getopts "t:" opt; do
case "$opt" in
  t) BUILD_TARGETS="$OPTARG" ;;
esac
done

mkdir -p $WORKDIR

git_latest_tag() {
  git -C "$1" describe --abbrev=0 --tags
}

git_checkout_reset() {
  local dir="$WORKDIR/$1" name="$1" repo="$2"
  if [ ! -d "$dir" ]; then
    (set -x;
      git clone "$repo" "$dir"
    )
  fi
  (set -x;
    git -C "$dir" fetch origin
  )
  local ver=$(git_latest_tag "$dir")
  echo "$name $ver"
  echo "$ver" > "$SRC/version.txt"
  (set -x;
    git -C "$dir" reset --hard
    git -C "$dir" clean -f -x -d -e node_modules
    git -C "$dir" checkout "$ver" &> /dev/null
  )
}

git_checkout_reset resvg "https://github.com/${REPO}.git"

declare -A TARGETS=(
  [darwin_amd64]=x86_64-apple-darwin
  [darwin_arm64]=aarch64-apple-darwin
  [linux_amd64]=x86_64-unknown-linux-gnu
  [linux_arm64]=aarch64-unknown-linux-gnu
  [linux_arm]=armv7-unknown-linux-gnueabihf
  [windows_amd64]=x86_64-pc-windows-gnu
)

#[dragonfly_amd64]=
#[freebsd_amd64]=x86_64-unknown-freebsd
#[netbsd_amd64]=x86_64-unknown-netbsd
#[openbsd_amd64]=
#[solaris_amd64]=x86_64-pc-solaris
#[windows_arm64]=aarch64-pc-windows-msvc

if [ -z "$BUILD_TARGETS" ]; then
  BUILD_TARGETS="${!TARGETS[@]}"
fi

BUILD_TARGETS=$(sed -e 's/\s\+/\n/g' <<< "$BUILD_TARGETS" | sort -i |tr '\n' ' ')

for f in $(find $WORKDIR/resvg -type f -name Cargo.toml); do
  cat > $(dirname "$f")/Cross.toml << __END__
[target.x86_64-apple-darwin]
image = "ghcr.io/cross-rs/x86_64-apple-darwin-cross:local"

[target.aarch64-apple-darwin]
image = "ghcr.io/cross-rs/aarch64-apple-darwin-cross:local"
__END__
done

mkdir -p $SRC/libresvg
(set -x;
  cp $WORKDIR/resvg/crates/c-api/resvg.h $SRC/libresvg
)

pushd $WORKDIR/resvg/crates/c-api &> /dev/null
for TARGET in $BUILD_TARGETS; do
  echo -e "\n\nBUILDING $TARGET"
  DEST=$SRC/libresvg/$TARGET
  mkdir -p $DEST
  RUST_TARGET="${TARGETS[$TARGET]}"
  if [ -z "$(rustup target list|grep "$RUST_TARGET"|grep installed)" ]; then
    (set -x;
      rustup target add $RUST_TARGET
    )
  fi
  (
    export "CARGO_TARGET_$(sed -e 's/-/_/g' <<< "$RUST_TARGET"|tr [:lower:] [:upper:])_RUSTFLAGS"="--print=native-static-libs"
    set -x
    cross build \
      --verbose \
      --release \
      --target $RUST_TARGET
    cp $WORKDIR/resvg/target/$RUST_TARGET/release/libresvg.a $DEST/libresvg.a
    chmod -x $DEST/libresvg.a
  )
done
popd &> /dev/null
