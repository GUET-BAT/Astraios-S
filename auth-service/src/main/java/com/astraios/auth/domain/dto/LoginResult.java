package com.astraios.auth.domain.dto;

import lombok.Data;

@Data
public class LoginResult {
    int status;
    String msg;
    String token;
}
