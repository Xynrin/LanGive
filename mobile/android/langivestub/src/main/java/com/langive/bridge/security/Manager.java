package com.langive.bridge.security;

public class Manager {
    public static Manager newSecurityManager() {
        return new Manager();
    }
    public void createPublicSession() {}
    public void createPrivateSession() {}
    public void joinSession(String sessionID, String deviceUUID) {}
    public void leaveSession(String sessionID, String deviceUUID) {}
}
