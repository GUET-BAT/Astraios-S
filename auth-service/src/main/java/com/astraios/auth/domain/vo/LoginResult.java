package com.astraios.auth.domain.vo;

import lombok.Data;

@Data
public class LoginResult {
    String accessToken;
    String refreshToken;
}
