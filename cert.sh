#!/bin/bash
if [ "$(expr substr $(uname -s) 1 5)" == "MINGW" ];then
    openssl req -x509 -nodes -days 365 -newkey rsa:4096 -keyout $1"test.key" -out $1"test.crt" -subj "//C=TG\ST=TG\L=TG\O=Dark Socket\OU=Dark Socket\CN=$2"
else 
    openssl req -x509 -nodes -days 365 -newkey rsa:4096 -keyout $1"test.key" -out $1"test.crt" -subj "/C=TG/ST=TG/L=TG/O=Dark Socket/OU=Dark Socket/CN=$2"
fi