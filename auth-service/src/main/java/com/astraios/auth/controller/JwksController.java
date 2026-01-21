package com.astraios.auth.controller;
import com.astraios.auth.service.AuthService;
import com.astraios.auth.utils.JwtTokenProvider;
import io.swagger.v3.oas.annotations.Operation;
import io.swagger.v3.oas.annotations.tags.Tag;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

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