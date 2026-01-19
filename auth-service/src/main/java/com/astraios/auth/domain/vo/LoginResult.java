package com.astraios.auth.domain.vo;

import lombok.Data;

@Data
public class LoginResult {
    int status;
    String msg;
    String accessToken;
    String refreshToken;
}
