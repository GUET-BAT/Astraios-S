package com.astraios.auth.domain.vo;

import lombok.Data;

@Data
public class LoginResult {
    int code;
    String msg;
    String accessToken;
    String refreshToken;
}
