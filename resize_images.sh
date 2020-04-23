#!/bin/bash
for f in feature_images/*
do
	if [ $(./magick identify -format "%w" $f) -gt 900 ]
	then
		./magick $f -resize '900' $f
	fi
done
