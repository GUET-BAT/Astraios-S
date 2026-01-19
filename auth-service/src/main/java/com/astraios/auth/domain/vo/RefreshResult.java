package com.astraios.auth.domain.vo;

import lombok.Data;

@Data
public class RefreshResult {
    int status;
    String msg;
    String accessToken;
    String refreshToken;
}
