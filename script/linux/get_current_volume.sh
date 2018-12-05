echo `amixer sget Master | awk -F"[][]" '/dB/ {print $2}' | awk -F"%" '{print $1}'`
