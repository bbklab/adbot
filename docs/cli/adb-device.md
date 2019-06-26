
# adbot adb-device

```bash
# adbot adb-device
NAME:
   adbot adb-device - adb device management

USAGE:
   adbot adb-device command [command options] [arguments...]

COMMANDS:
     watch          watch all of adb devices
     ls             list adb devices
     inspect        inspect details of an adb device
     screencap      take screencap on an adb device
     dumpui         dump ui nodes on an adb device
     click          click adb device's UI Coordinate
     goback         tap adb device back key
     gotohome       tap adb device home key
     reboot         reboot an adb device
     exec           exec command on an adb device
     set-bill       set abb device max bill perday, must between [0-10000], 0 means unlimited
     set-amount     set abb device max amount perday, by CNY, must between [0-100000000], 0 means unlimited
     set-weight     set adb device weight value, must between [0-100], the higher value means the higher weight, 0 means disabled
     bind-alipay    bind abb device with alipay account
     revoke-alipay  revoke abb device alipay account
     rm             remove abb device
```
