
.PHONY: pdf2thumb lambda_x64.zip lambda_arm64.zip

all: pdf2thumb lambda_x64.zip lambda_arm64.zip

pdfium-linux-x64.tgz:
	curl -LO https://github.com/bblanchon/pdfium-binaries/releases/latest/download/pdfium-linux-x64.tgz

pdfium-linux-arm64.tgz:
	curl -LO https://github.com/bblanchon/pdfium-binaries/releases/latest/download/pdfium-linux-arm64.tgz

pdfium-mac-x64.tgz:
	curl -LO https://github.com/bblanchon/pdfium-binaries/releases/latest/download/pdfium-mac-x64.tgz

.linux_x64: pdfium-linux-x64.tgz
	rm -fR _linux_x64
	mkdir _linux_x64
	(cd _linux_x64 && tar zxf ../pdfium-linux-x64.tgz && echo prefix=`pwd` > pdfium.pc )
	cat pdfium_tpl.pc >> _linux_x64/pdfium.pc
	
.linux_arm64: pdfium-linux-arm64.tgz
	rm -fR _linux_arm64
	mkdir _linux_arm64
	(cd _linux_arm64 && tar zxf ../pdfium-linux-arm64.tgz && echo prefix=`pwd` > pdfium.pc )
	cat pdfium_tpl.pc >> _linux_arm64/pdfium.pc

.macos_x64: pdfium-mac-x64.tgz
	rm -fR _macos_x64
	mkdir _macos_x64
	(cd _macos_x64 && tar zxf ../pdfium-mac-x64.tgz && echo prefix=`pwd` > pdfium.pc )
	cat pdfium_tpl.pc >> _macos_x64/pdfium.pc

pdf2thumb: .macos_x64
	CGO_ENABLED=1 \
	PKG_CONFIG_PATH=`pwd`/_macos_x64 \
	go build -o pdf2thumb
	cp _macos_x64/lib/libpdfium.dylib .

lambda_x64.zip: .linux_x64
	GOOS=linux \
	CGO_ENABLED=1 \
	CC=`pwd`/zcc.sh \
	CXX=`pwd`/zxx.sh \
	PKG_CONFIG_PATH=`pwd`/_linux_x64 \
	GOARCH=amd64 \
	ZTARGET=x86_64-linux-gnu \
	go build -tags lambda -ldflags="-linkmode external" -o _linux_x64/pdf2thumb
	(cd _linux_x64 && cp ../bootstrap lib/*.so . && zip ../lambda_x64.zip bootstrap pdf2thumb libpdfium.so)

lambda_arm64.zip: .linux_arm64
	GOOS=linux \
	CGO_ENABLED=1 \
	CC=`pwd`/zcc.sh \
	CXX=`pwd`/zxx.sh \
	PKG_CONFIG_PATH=`pwd`/_linux_arm64 \
	GOARCH=arm64 \
	ZTARGET=aarch64-linux-gnu \
	go build -tags lambda -ldflags="-linkmode external" -o _linux_arm64/pdf2thumb
	(cd _linux_arm64 && cp ../bootstrap lib/*.so . && zip ../lambda_arm64.zip bootstrap pdf2thumb libpdfium.so)

clean:
	rm -fR _linux_x64 _linux_arm64 _macos_x64 lambda_x64.zip lambda_arm64.zip pdf2thumb libpdfium.dylib pdfium-*.tgz
