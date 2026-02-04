package com.astraios.auth.domain.vo;

import lombok.Data;

@Data
public class RefreshResult {
    String accessToken;
    String refreshToken;
}
