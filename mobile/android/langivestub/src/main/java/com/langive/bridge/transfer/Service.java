package com.langive.bridge.transfer;

public class Service {
    public void start() throws Exception {}
    public void stop() {}
    public void setOnIncomingRequest(IncomingRequestHandler handler) {}
    public String approveIncoming(String id) throws Exception { return ""; }
    public void rejectIncoming(String id) throws Exception {}
    public IncomingRequestSlice pendingRequests() { return new IncomingRequestSlice(); }
    public TransferStatusSlice getTransfers() { return new TransferStatusSlice(); }
    public void cancelTransfer(String id) throws Exception {}
    public void clearCompleted() {}
    public void sendFilesAs(String address, String fromName, StringSlice files) throws Exception {}
    public void sendFolderAs(String address, String fromName, String folderPath) throws Exception {}
}
