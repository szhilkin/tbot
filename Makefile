upload:
	rsync . pi@192.168.88.94:/home/pi/telegram-bot

run:
	ssh -p 22 pi@192.168.88.94 "cd telegram-bot && go build main.go && sudo nohup ./main &"

deploy: upload run
