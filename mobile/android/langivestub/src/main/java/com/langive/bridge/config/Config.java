package com.langive.bridge.config;

public class Config {
    public static Config load() throws Exception {
        return new Config();
    }

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
    public void setDeviceName(String name) { this.deviceName = name; }

    public String getDeviceUUID() { return deviceUUID; }
    public void setDeviceUUID(String uuid) { this.deviceUUID = uuid; }

    public String getDeviceToken() { return deviceToken; }
    public void setDeviceToken(String token) { this.deviceToken = token; }

    public String getDownloadPath() { return downloadPath; }
    public void setDownloadPath(String path) { this.downloadPath = path; }

    public long getPort() { return port; }
    public void setPort(long port) { this.port = port; }

    public boolean getPrivacyMode() { return privacyMode; }
    public void setPrivacyMode(boolean privacyMode) { this.privacyMode = privacyMode; }

    public String getSessionID() { return sessionID; }
    public void setSessionID(String id) { this.sessionID = id; }

    public String getVersion() { return version; }
    public void setVersion(String version) { this.version = version; }

    public boolean getAutoUpdate() { return autoUpdate; }
    public void setAutoUpdate(boolean autoUpdate) { this.autoUpdate = autoUpdate; }

    public long getScanInterval() { return scanInterval; }
    public void setScanInterval(long scanInterval) { this.scanInterval = scanInterval; }

    public boolean getFirstRun() { return firstRun; }
    public void setFirstRun(boolean firstRun) { this.firstRun = firstRun; }

    public boolean getBackgroundMode() { return backgroundMode; }
    public void setBackgroundMode(boolean backgroundMode) { this.backgroundMode = backgroundMode; }

    public void save() throws Exception {}
    public long getDeviceTimeout() { return scanInterval * 3; }
}
