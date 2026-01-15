package com.astraios.auth.domain.dto;

import lombok.Data;

@Data
public class RegisterRequest {
    String username;
    String password;
}
