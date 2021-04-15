# Configuration
Please configuration the device instance and device model. You could refer to the example in build/crd-samples/devices/. 
# Notes
- Please configure the remote certification file if you use certification&key authentication.
- The format of all return values is a string.
- Not all value types are support now. The supported types include:
  Boolean
  String
  ByteString
  XMLElement
  LocalizedText
  QualifiedName
  SignedByte
  Int16
  Int32
  Int64
  Byte
  Uint16
  Uint32
  Uint64
  Float
  Double
- The get device status function "driver.GetStatus" should be written depending the device.

# Notice

There's a bug about gopcua library. https://github.com/gopcua/opcua/issues/410.
If you want to use username&password mode, please modify the file as this:
vendor/github.com/gopcua/opcua/uasc/secure_channel_crypto.go line 67:

```
 67 func (s *SecureChannel) EncryptUserPassword(policyURI, password string, cert, nonce []byte) ([]byte, string, error) {
 68     // If the User ID Token's policy was null, then default to the secure channel's policy
 69     if policyURI == "" {
 70         policyURI = s.cfg.SecurityPolicyURI
 71     }
 72
 73     if policyURI == ua.SecurityPolicyURINone {
 74         return []byte(password), "", nil
 75     }
```
