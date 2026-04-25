#/bin/sh

#for i in *.avi; do ffmpeg -i "$i" "${i%.*}.mp4"; done

for i in *.mp4; do ffmpeg -i "$i" -vf "fps=20,scale=120:240:flags=lanczos,eq=saturation=2.5:gamma=0.8" -q:v 9 "${i%.*}_20fps.mjpeg"; done