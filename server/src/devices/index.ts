import WebSocket from 'ws';
import {
  AnyDeviceData,
  DateTime,
  Device as DeviceType,
} from '@gbaranski/types';
import { logSocketError } from '@/cli';
import WatermixerDevice from './watermixer';
import AlarmclockDevice from './alarmclock';

export type AnyDeviceObject = WatermixerDevice | AlarmclockDevice;

export default abstract class Device<DeviceData extends AnyDeviceData> {
  private static _currentDevices: AnyDeviceObject[] = [];

  public static get currentDevices(): AnyDeviceObject[] {
    return this._currentDevices;
  }

  public static addNewDevice(device: AnyDeviceObject): void {
    this._currentDevices.push(device);
  }

  public static removeDevice(device: AnyDeviceObject): void {
    this._currentDevices = this._currentDevices.filter(
      (_device: AnyDeviceObject) => _device !== device,
    );
  }

  protected static updateDevice(
    deviceUid: string,
    deviceData: AnyDeviceData,
  ): void {
    this._currentDevices.map(_device =>
      _device.deviceUid === deviceUid ? deviceData : _device.deviceData,
    );
  }

  private _status = false;

  constructor(
    protected ws: WebSocket,
    private _deviceData: DeviceData,
    public readonly deviceType: DeviceType.DeviceType,
    public readonly deviceUid: string,
  ) {
    this._status = true;
  }

  abstract handleMessage(message: WebSocket.Data): void;

  public requestDevice(
    type: DeviceType.RequestType,
    data?: DateTime | boolean,
  ): boolean {
    if (!this.ws) {
      throw new Error('Websocket is not defined');
    }
    if (!this.ws.OPEN) {
      throw new Error('Websocket is not at OPEN state');
    }
    if (!this._status) {
      throw new Error('Device status is false');
    }
    const requestData = {
      type,
      data,
    };
    console.log(requestData);
    this.ws.send(JSON.stringify(requestData));

    return true;
  }

  public terminateConnection(reason: string): void {
    this.ws.terminate();
    logSocketError(this.deviceType, this.deviceUid, reason, 'device');
  }

  get deviceData(): DeviceData {
    return this._deviceData;
  }

  set deviceData(data: DeviceData) {
    this._deviceData = data;
  }

  set status(status: boolean) {
    this._status = status;
  }

  get status(): boolean {
    return this._status;
  }
}