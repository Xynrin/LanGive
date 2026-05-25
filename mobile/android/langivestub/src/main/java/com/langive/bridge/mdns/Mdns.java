package com.langive.bridge.mdns;

public class Mdns {
    public static Service newService(String deviceName, String deviceUUID, long port, String version, String sessionID, boolean privacy) {
        return new Service();
    }
}
