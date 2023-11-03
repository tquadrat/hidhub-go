#! /bin/sh

go build
if [ $? -eq 0 ]
then
    ./hidhub-go -vendorId 0x046a -productId 0x0023
fi
