package com.astraios.auth.domain.dto;


import lombok.Data;

@Data
public class LoginRequest {
    String username;
    String password;
}