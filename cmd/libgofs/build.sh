#/usr/bin/env sh

LIBPATH=$(pwd)
if [ $(echo ${CURRENT_PATH} | grep "cmd/libgofs$" | wc -l) -eq 0 ]; then
    LIBPATH="${LIBPATH}/cmd/libgofs"
fi

BUILDPATH="${LIBPATH}/build"
ROOTBIN="${LIBPATH}/../../bin"
mkdir -p ${BUILDPATH} ${ROOTBIN}

echo "Building C-Shared lib from Go code"
go build -o ${BUILDPATH}/libgofs.so -buildmode=c-shared ${LIBPATH}/libgofs.go
cp ${BUILDPATH}/libgofs.h ${LIBPATH}

echo "Building C libs"
gcc -g -o ${BUILDPATH}/ssl.o -c ${LIBPATH}/ssl.c
gcc -g -o ${BUILDPATH}/ret.o -c ${LIBPATH}/ret.c

echo "Building C main"
gcc ${LIBPATH}/maingofs.c -o ${ROOTBIN}/maingofs ${BUILDPATH}/ret.o ${BUILDPATH}/ssl.o ${BUILDPATH}/libgofs.so
