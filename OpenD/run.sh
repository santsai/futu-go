#!/bin/bash

CFG=opend-dev.xml
PEM=opend-dev-key.pem

if [ $# -gt 0 ]; then
    CFG=$1
    shift
fi

if [ $# -gt 0 ]; then
    PEM=$1
    shift
fi

BASE=/root/.com.futunn.FutuOpenD
cp $BASE/$CFG /opend/FutuOpenD.xml
cp $BASE/$PEM /opend/key.pem

/opend/FutuOpenD
