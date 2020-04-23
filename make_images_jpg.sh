#!/bin/bash
for f in feature_images/*
do
	if [ $(./magick identify -format "%B" $f) -gt 1000000 ]
	then
		./magick $f "${f}.jpg"
	fi
done
