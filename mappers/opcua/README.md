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
