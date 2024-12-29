#!/usr/bin/bash
echo -e "BEWARE:"
echo -e "First remove else:~/public with root-permissions."
echo -e "After this you can use this script."
echo -e ""
rsync -rav ./public/ else:~/public/
