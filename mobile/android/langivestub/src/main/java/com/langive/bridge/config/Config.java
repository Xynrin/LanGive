package com.langive.bridge.config;

public class Config {
    public static Settings load() throws Exception {
        return new Settings();
    }

    public static String getDefaultDownloadPath() { return "/sdcard/Download"; }
    public static String getDefaultDeviceName() { return "Android Device"; }
    public static String currentVersion() { return "1.0.0"; }
}
