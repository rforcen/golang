ld -r -b binary -o presets.o *.vsl
nm presets.o
nm -P presets.o | grep _size

