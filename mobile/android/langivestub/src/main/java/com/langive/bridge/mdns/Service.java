package com.langive.bridge.mdns;

public class Service {
    public void start() throws Exception {}
    public void stop() {}
    public void setDeviceName(String name) {}
    public void setPrivacy(boolean enabled) {}
    public void setSession(String sessionID) {}
    public DeviceInfoSlice getDiscoveredDevices() { return new DeviceInfoSlice(); }
    public DeviceInfoSlice getPublicDevices() { return new DeviceInfoSlice(); }
    public DeviceInfo getDevice(String uuid) { return null; }
    public void startCleanupRoutine(long interval, long timeout) {}
    public void removeStaleDevices(long timeout) {}
}
