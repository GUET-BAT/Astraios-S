package com.astraios.auth;

import com.astraios.auth.utils.JwtTokenProvider;
import io.jsonwebtoken.ExpiredJwtException;
import io.jsonwebtoken.Jwts;
import org.junit.jupiter.api.Test;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.test.context.SpringBootTest;

import java.security.KeyPair;
import java.util.Date;

@SpringBootTest
class AuthApplicationTests {

    @Autowired
    private JwtTokenProvider jwtTokenProvider;

	@Test
	void contextLoads() {
	}
	@Test
	public void testExpiredToken() throws InterruptedException {

		// 1. 生成一个 1 秒后过期的 Token
		String token = jwtTokenProvider.generateAccessToken("123", "123");

		// 2. 此时解析是正常的
		System.out.println("First parse: " + jwtTokenProvider.parseToken(token));

		// 3. 等待 2 秒，让它过期
		Thread.sleep(2000);

		// 4. 再次解析，这里必然抛出 ExpiredJwtException
		try {
			jwtTokenProvider.parseToken(token);
		} catch (ExpiredJwtException e) {
			System.out.println("Caught Expected Exception: " + e.getMessage());
			// 输出通常包含: "JWT expired at 2023-xx-xx..."
		}
	}
}


