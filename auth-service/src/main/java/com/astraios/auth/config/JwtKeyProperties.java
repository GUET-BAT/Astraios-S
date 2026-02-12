package com.astraios.auth.config;
 
import lombok.Data;
import org.springframework.boot.context.properties.ConfigurationProperties;
import org.springframework.stereotype.Component;
import org.springframework.util.StringUtils;
import java.util.UUID;
 
@Component
@ConfigurationProperties(prefix = "jwt")
@Data
public class JwtKeyProperties {
     /**
      * RSA public key in PEM format.
      */
     private String publicKey;
     /**
      * RSA private key in PEM (PKCS#8) format.
      */
     private String privateKey;
     /**
      * Optional key id (kid) for JWK/JWT headers.
      */
     private String keyId;
 
     public String getKeyIdOrDefault() {
         return StringUtils.hasText(keyId) ? keyId : UUID.randomUUID().toString();
     }
}
