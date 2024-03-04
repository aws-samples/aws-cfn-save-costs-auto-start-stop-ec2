# ec2sched

Containing lambda code for AutoStart/AutoStop, StartWeekDay/StopWeekDay

Asgsched allows for customizable Start Stop schedules to fit a variaty of use cases. The code evaluates tags on the autoscaling groups determining actions.

Code Info
+ Code base language [go1.20.1](https://go.dev/doc/)
<br />
<br />

## Schedule Functions

The sheduling lambda functions operate by in large the same way. Collects the autoscaling groups, views the corresponding tag, and takes actions based on the tag value.
> [!NOTE]
> Tag names and values are case sensitive.

If action is needed the autoscaling group actions are suspended (Launch, Terminate, Alarm Notification, etc.) then the underlying instances are stopped, the inverse happens when autoscaling groups need to start: underlying autoscaling group instance are started, then autoscaling actions are resumed. Fuction specifics outlined below

<br />
<br />

### AutoStart/AutoStop

AutoStart/AutoStop functions look for the AutoSart/AutoStop tag on the autoscaling group. If the tag (AutoStart/AutoStop) exists and value is `true` the function takes action.

The deployed lambda for AutoStart and AutoStop is typically triggered via eventbridge rules running at the desired times

Deploy Specifics
+ Tag Names
    - AutoStart
    - AutoStop
+ Tag Values
    - `true`
    - If that tag value is blank or a value other than `true` no action will take place.

<br />
<br />

### Start/Stop and WeekDay/WeekEnd

Start/Stop WeekDay/WeekEnd functions look for time values via tags on the autoscaling group. If that tag is present and the value (time) is plus or minus 5 minutes of the function running, currently a weekday is (Monday-Friday according to TZ tag), the start stop action will take place.

The deployed lambda for Start/Stop WeekDay/WeekEnd is intended to triggered via eventbridge rule which runs every 5min.

Deploy Specifics
+ Tag Names
    - startweekday
    - stopweekday
    - startweekend
    - stopweekend
+ Tag Values
    - 24 hour values i.e. 14:00 for 2pm
+ Environment Variables
    - TZ: Timezone the specified time value will be assessed i.e. US/Pacific, US/Eastern, Europe/London **Required**

---
<br />

# Build Binaries

There are different options for OS and architecture (GOOS & GOARCH). To build on linux for arm64
`env GOOS=linux GOARCH=arm64 go build -o bootstrap main.go`
> [!NOTE]
> Be sure to set the proper lambda properties i.e. Runtime, Architectures, and Handler that work with your GOOS and GOARCH values.

