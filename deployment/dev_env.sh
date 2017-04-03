#!/usr/bin/env bash

# Fix auto=completion of PyCharm
# Adding symbolic links in the vendor map own packages
rm -R src/vendor/github.com/lukin0110/push.kiwi/ || true
mkdir -p src/vendor/github.com/lukin0110/push.kiwi/
ln -s "$(pwd)/src/sanitize/" src/vendor/github.com/lukin0110/push.kiwi/sanitize
ln -s "$(pwd)/src/utils/" src/vendor/github.com/lukin0110/push.kiwi/utils
