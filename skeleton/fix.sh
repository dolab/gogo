#!/usr/bin/env bash

APP="${PWD##*/}"
if [ ${APP} != "skeleton" ]; then
    echo "Replacing skeleton with ${APP} ..."
    grep -rl skeleton * | xargs sed -i "" "s/skeleton/${APP}/g"

    echo "Done!"
    echo
    echo "Now you can run source env.sh && make godev"
    rm ./fix.sh
fi
