import 'package:homeflow/models/device.dart';
import 'package:homeflow/models/devices/watermixer.dart';
import 'package:flutter/material.dart';
import 'package:homeflow/services/firebase.dart';
import 'package:mqtt_client/mqtt_client.dart';
import 'package:homeflow/services/mqtt.dart';
import 'package:homeflow/utils/misc.dart';
import 'package:homeflow/shared/constants.dart';
import 'dart:async';

const tenMinutesInMillis = 1000 * 10 * 60;

class Watermixer extends StatefulWidget {
  final FirebaseDevice firebaseDevice;

  Watermixer({@required this.firebaseDevice});

  @override
  _WatermixerState createState() => _WatermixerState();
}

class _WatermixerState extends State<Watermixer> {
  Timer _countdownTimer;
  String mixingTimeString = "";

  void startCounting(int finishMixTimestamp) {
    final callback = (Timer timer) {
      if (!this.mounted) {
        timer.cancel();
        return;
      }
      setState(() {
        mixingTimeString =
            durationToString(getEpochDiffDuration(finishMixTimestamp));
      });
    };

    _countdownTimer = Timer.periodic(Duration(seconds: 1), callback);
    callback(_countdownTimer);
  }

  @override
  void dispose() {
    _countdownTimer.cancel();
    _countdownTimer = null;
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final WatermixerData data =
        WatermixerData.fromJson(widget.firebaseDevice.data);
    startCounting(data.finishMixTimestamp);

    final startMixing = () async {
      print("MQTT CONN STAT: ${MqttService.mqttClient.connectionStatus}");
      final String uid = widget.firebaseDevice.uid;
      final RequestTopic topic = RequestTopic(
          request: '$uid/event/startmix/request',
          response: '$uid/event/startmix/response');

      bool hasCompleted = false;
      final Future req = MqttService.sendMessage(
          topic: topic, qos: MqttQos.atMostOnce, data: null);

      req.whenComplete(() {
        hasCompleted = true;
        const snackbar = SnackBar(
          content: Text("Success mixing water!"),
          duration: Duration(milliseconds: 500),
        );
        Scaffold.of(context).showSnackBar(snackbar);
        final WatermixerData newDeviceData = WatermixerData(
            finishMixTimestamp:
                DateTime.now().millisecondsSinceEpoch + tenMinutesInMillis);
        final FirebaseDevice newDevice = FirebaseDevice(
            data: newDeviceData.toJson(),
            ip: widget.firebaseDevice.ip,
            status: widget.firebaseDevice.status,
            type: widget.firebaseDevice.type,
            uid: widget.firebaseDevice.uid);
        FirebaseService.updateFirebaseDevice(newDevice);
      });
      Future.delayed(Duration(seconds: 2), () {
        if (!hasCompleted) {
          const snackbar = SnackBar(content: Text("No response from device!"));
          Scaffold.of(context).showSnackBar(snackbar);
        }
      });
    };

    return ConstrainedBox(
      constraints: BoxConstraints(minHeight: CardMinHeight),
      child: Card(
          child: InkWell(
        splashColor: Colors.blue.withAlpha(30),
        onTap: () {
          print('Card tapped.');
        },
        child: Container(
          child: Column(children: [
            SizedBox(
              height: 5,
            ),
            Text("Watermixer", style: TextStyle(fontSize: 24)),
            Divider(
              indent: 20,
              endIndent: 20,
              thickness: 1,
            ),
            Row(mainAxisAlignment: MainAxisAlignment.spaceEvenly, children: [
              Column(children: [
                Text(
                  "Mixing state",
                  style: TextStyle(
                      fontWeight: FontWeight.w300,
                      fontSize: 14,
                      color: Colors.black.withOpacity(0.6)),
                ),
                Text(
                    data.finishMixTimestamp >
                            DateTime.now().millisecondsSinceEpoch
                        ? "Mixing!"
                        : "Idle",
                    style:
                        TextStyle(fontSize: 26, fontWeight: FontWeight.w300)),
              ]),
              Column(children: [
                Text(
                  "Mixing time",
                  style: TextStyle(
                      fontWeight: FontWeight.w300,
                      fontSize: 14,
                      color: Colors.black.withOpacity(0.6)),
                ),
                Text(mixingTimeString,
                    style:
                        TextStyle(fontSize: 26, fontWeight: FontWeight.w300)),
              ]),
            ]),
            Row(
              mainAxisAlignment: MainAxisAlignment.center,
              children: [
                ButtonBar(
                  alignment: MainAxisAlignment.center,
                  children: <Widget>[
                    FlatButton(
                      child: const Text('START MIXING'),
                      onPressed: () {
                        startMixing();
                      },
                    ),
                  ],
                ),
              ],
            )
          ]),
        ),
      )),
    );
  }
}
