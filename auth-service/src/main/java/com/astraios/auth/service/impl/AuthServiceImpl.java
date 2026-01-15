package com.astraios.auth.service.impl;

import com.astraios.auth.domain.dto.LoginRequest;
import com.astraios.auth.domain.dto.LoginResult;
import com.astraios.auth.domain.dto.RegisterRequest;
import com.astraios.auth.service.AuthService;
import com.astraios.auth.utils.JwtTokenProvider;
import lombok.RequiredArgsConstructor;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.security.authentication.AuthenticationManager;
import org.springframework.security.authentication.UsernamePasswordAuthenticationToken;
import org.springframework.security.core.Authentication;
import org.springframework.security.core.userdetails.UserDetails;
import org.springframework.security.crypto.password.PasswordEncoder;
import org.springframework.stereotype.Service;
import org.springframework.util.StringUtils;


@Service
@RequiredArgsConstructor
public class AuthServiceImpl implements AuthService {

    private final AuthenticationManager authenticationManager;
    private final PasswordEncoder passwordEncoder;

    @Override
    public LoginResult login(LoginRequest request) {
        try {
            LoginResult loginResult = new LoginResult();

            // 1.校验请求
            if(!StringUtils.hasText(request.getUsername()) && !StringUtils.hasText(request.getPassword())){
                loginResult.setMsg("Invalid username or password");
                loginResult.setStatus(HttpStatus.UNAUTHORIZED.value());
                return loginResult;
            }

            // 2. 封装认证请求
            Authentication requestAuth = new UsernamePasswordAuthenticationToken(request.getUsername(), request.getPassword());

            // 3.执行认证
            Authentication auth = authenticationManager.authenticate(requestAuth);

            // 4.获取认证结果
            UserDetails details = (UserDetails) auth.getPrincipal();

            //5.生成token并返回
            String token = JwtTokenProvider.generateToken(details.getUsername());
            loginResult.setToken(token);
            loginResult.setStatus(HttpStatus.OK.value());
            loginResult.setMsg("Successfully logged in");
            return loginResult;
        }
        catch (Exception e) {
            LoginResult loginResult = new LoginResult();
            loginResult.setMsg(e.getMessage());
            loginResult.setStatus(HttpStatus.UNAUTHORIZED.value());
            return loginResult;
        }
    }

    @Override
    public ResponseEntity<?> register(RegisterRequest request) {
        try {

            // 1.校验请求
            if(!StringUtils.hasText(request.getUsername()) && !StringUtils.hasText(request.getPassword())){
                return ResponseEntity.status(HttpStatus.UNAUTHORIZED).body("Invalid username or password");
            }

            // 2. 密码加密
            String encodedPassword = passwordEncoder.encode(request.getPassword());

            // 3. TODO RPC调用user服务注册

            return ResponseEntity.status(HttpStatus.CREATED).body("success");
        }
        catch (Exception e) {

            return ResponseEntity.status(HttpStatus.UNAUTHORIZED).body("Invalid username or password");
        }
    }
}
