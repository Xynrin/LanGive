package com.langive.bridge.config;

public class Settings {
    private String deviceName = "Android Device";
    private String deviceUUID = "stub-uuid";
    private String deviceToken = "stub-token";
    private String downloadPath = "/sdcard/Download";
    private long port = 5566;
    private boolean privacyMode = false;
    private String sessionID = "public";
    private String version = "1.0.0";
    private boolean autoUpdate = false;
    private long scanInterval = 5;
    private boolean firstRun = false;
    private boolean backgroundMode = false;

    public String getDeviceName() { return deviceName; }
    public void setDeviceName(String v) { this.deviceName = v; }

    public String getDeviceUUID() { return deviceUUID; }
    public void setDeviceUUID(String v) { this.deviceUUID = v; }

    public String getDeviceToken() { return deviceToken; }
    public void setDeviceToken(String v) { this.deviceToken = v; }

    public String getDownloadPath() { return downloadPath; }
    public void setDownloadPath(String v) { this.downloadPath = v; }

    public long getPort() { return port; }
    public void setPort(long v) { this.port = v; }

    public boolean getPrivacyMode() { return privacyMode; }
    public void setPrivacyMode(boolean v) { this.privacyMode = v; }

    public String getSessionID() { return sessionID; }
    public void setSessionID(String v) { this.sessionID = v; }

    public String getVersion() { return version; }
    public void setVersion(String v) { this.version = v; }

    public boolean getAutoUpdate() { return autoUpdate; }
    public void setAutoUpdate(boolean v) { this.autoUpdate = v; }

    public long getScanInterval() { return scanInterval; }
    public void setScanInterval(long v) { this.scanInterval = v; }

    public boolean getFirstRun() { return firstRun; }
    public void setFirstRun(boolean v) { this.firstRun = v; }

    public boolean getBackgroundMode() { return backgroundMode; }
    public void setBackgroundMode(boolean v) { this.backgroundMode = v; }

    public void save() throws Exception {}
    public long getDeviceTimeout() { return scanInterval * 3; }
}
