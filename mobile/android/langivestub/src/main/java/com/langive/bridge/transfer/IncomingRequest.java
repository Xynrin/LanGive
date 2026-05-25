package com.langive.bridge.transfer;

public class IncomingRequest {
    private String id = "";
    private String fromName = "";
    private String fromAddr = "";
    private String fileName = "";
    private long totalSize = 0;
    private long receivedAt = 0;

    public String getId() { return id; }
    public String getFromName() { return fromName; }
    public String getFromAddr() { return fromAddr; }
    public String getFileName() { return fileName; }
    public long getTotalSize() { return totalSize; }
    public long getReceivedAt() { return receivedAt; }
}
