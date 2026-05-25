package com.langive.bridge.mdns;

public class DeviceInfo {
    private String id = "";
    private String name = "";
    private String address = "";
    private long port = 5566;
    private String platform = "";
    private String uuid = "";
    private String version = "";
    private String sessionID = "";
    private boolean isPublic = false;
    private boolean privacy = false;
    private long lastSeen = 0;

    public String getId() { return id; }
    public String getName() { return name; }
    public String getAddress() { return address; }
    public long getPort() { return port; }
    public String getPlatform() { return platform; }
    public String getUuid() { return uuid; }
    public String getVersion() { return version; }
    public String getSessionID() { return sessionID; }
    public boolean getIsPublic() { return isPublic; }
    public boolean getPrivacy() { return privacy; }
    public long getLastSeen() { return lastSeen; }
}
