# Screen time-lapse
Simple util to make time-lapse from your screen.

## Install
```
git clone https://github.com/fullpipe/screen-time-lapse
cd screen-time-lapse && make && sudo cp ./bin/screen-time-lapse /usr/local/bin/
```
## Usage
Make a screenshot every second and save to `./today` directory.
```
screen-time-lapse 1 ./today
```
![systray](systray.png)

### Todo
* make it work not only in unix systems
* add image sequence to video converter
* add image sequence to gif converter