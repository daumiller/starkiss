ROKU_ADDR=192.168.0.199
TVU_ADDR=192.168.0.182
TVD_ADDR=192.168.0.186
ROKU_USER=rokudev
ROKU_PASS=rokudev
PORT_CONS=8085
PORT_DEBG=8080
ZIP_FILE=starkiss.zip
CONTENTS=manifest source components images sounds fonts

all: install

install: zip
	@curl --anyauth -u $(ROKU_USER):$(ROKU_PASS) -s -S -F "mysubmit=Install" -F "archive=@$(ZIP_FILE)" -F "passwd=" http://$(ROKU_ADDR)/plugin_install \
		| grep "<font color"             \
		| sed "s/<font color=\"[^\"]*\">//" \
		| sed "s/<\/font>//"

installtvu: zip
	@curl --anyauth -u $(ROKU_USER):$(ROKU_PASS) -s -S -F "mysubmit=Install" -F "archive=@$(ZIP_FILE)" -F "passwd=" http://$(TVU_ADDR)/plugin_install \
		| grep "<font color"             \
		| sed "s/<font color=\"[^\"]*\">//" \
		| sed "s/<\/font>//"

installtvd: zip
	@curl --anyauth -u $(ROKU_USER):$(ROKU_PASS) -s -S -F "mysubmit=Install" -F "archive=@$(ZIP_FILE)" -F "passwd=" http://$(TVD_ADDR)/plugin_install \
		| grep "<font color"             \
		| sed "s/<font color=\"[^\"]*\">//" \
		| sed "s/<\/font>//"


zip:
	@rm -f $(ZIP_FILE)
	@zip -9 -r $(ZIP_FILE) $(CONTENTS)

console:
	telnet $(ROKU_ADDR) $(PORT_CONS)

debug:
	telnet $(ROKU_ADDR) $(PORT_DEBG)
