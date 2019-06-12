#!/bin/bash

adb devices -l

devices=$(adb devices | tail -n +2 | awk '{print $1}')
for device in ${devices}
do
	echo "=========$device========="
	# adb -s $device shell echo $EXTERNAL_STORAGE  # not working
	# adb -s $device shell dumpsys notification
	adb -s $device shell dumpsys notification | grep "tickerText"
	# adb -s $device shell getprop  wifi.interface
	# adb -s $device shell input keyevent 26     # pwoer switch
	# adb -s $device shell input keyevent 223    # light off

	# ensure screen light on
	adb -s $device shell input keyevent 224            # light on
	if !( adb -s $device shell dumpsys window policy   | grep -q "mScreenOnEarly=true" ); then    # if screen still power off
		adb -s $device shell input keyevent 26         # power switch
	fi

	# ensure @ homepage
	adb -s $device shell input swipe 300 1000 300 500  # WIPE UP to unlock
	adb -s $device shell input keyevent 3              # HOME
	# xmlf=$(adb -s $device shell uiautomator dump | awk -F "dumped to:" '(/dumped to:/){print $NF;exit}' | tr -d '\r\n')
	# lxmlf=${xmlf##*/}.${device}
	# adb -s $device pull $xmlf ${lxmlf}
	# grep -E -o "text=[^ ]*" ${lxmlf} |grep -v -E "\"\""

	# ensure sms page
    # 546052d21f384 --> bounds="[238,1230][310,1266]
	# cX=$((238/2+310/2))
	# cY=$((1230/2+1266/2))
	# adb -s $device shell input touchscreen tap $cX $cY

	# pull down
	adb -s $device shell input swipe 300 0 300 1000 # pull down from top

done




# ---> tail -f andriod log
# adb -s 546052d21f384 logcat  -c
# adb -s 546052d21f384 logcat  -d | wc -l
# 05-20 21:23:06.612  1637  1637 D PhoneStatusBar: addNotification pkg=com.eg.android.AlipayGphone;basepkg=com.eg.android.AlipayGphone;id=-469726375
