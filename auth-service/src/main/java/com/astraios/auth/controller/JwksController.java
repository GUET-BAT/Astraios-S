package com.astraios.auth.controller;
import com.astraios.auth.utils.JwtTokenProvider;
import lombok.RequiredArgsConstructor;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RestController;

import java.util.Map;

@RestController
@RequiredArgsConstructor
public class JwksController {


    private final JwtTokenProvider jwtTokenProvider;


    @GetMapping("/.well-known/jwks.json")
    public Map<String, Object> keys() {
        return jwtTokenProvider.getJwkSet();
    }
}
