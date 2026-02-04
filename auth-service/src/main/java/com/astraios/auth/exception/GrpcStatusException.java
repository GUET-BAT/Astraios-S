package com.astraios.auth.exception;

import io.grpc.Status;

public class GrpcStatusException extends RuntimeException {
    private final Status status;

    public GrpcStatusException(Status status, String message) {
        super(message);
        this.status = status;
    }

    public GrpcStatusException(Status status, String message, Throwable cause) {
        super(message, cause);
        this.status = status;
    }

    public Status getStatus() {
        return status;
    }
}
