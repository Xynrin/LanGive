package com.langive.bridge.transfer;

public class TransferStatus {
    private String id = "";
    private String type = "";
    private String fileName = "";
    private long totalSize = 0;
    private long sentSize = 0;
    private double progress = 0.0;
    private String status = "";
    private String error = "";
    private String peerAddr = "";

    public String getId() { return id; }
    public String getType() { return type; }
    public String getFileName() { return fileName; }
    public long getTotalSize() { return totalSize; }
    public long getSentSize() { return sentSize; }
    public double getProgress() { return progress; }
    public String getStatus() { return status; }
    public String getError() { return error; }
    public String getPeerAddr() { return peerAddr; }
}
