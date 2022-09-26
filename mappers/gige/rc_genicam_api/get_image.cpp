#include "get_image.h"


class IOException : public std::exception {
public:

	IOException(const std::string& _msg) { msg = _msg; }
	virtual const char* what() const noexcept { return msg.c_str(); }

private:

	std::string msg;
};

unsigned int storeBuffer(rcg::ImgFmt fmt, const rcg::Buffer* buffer, unsigned char** image_buffer, uint32_t part, size_t yoffset = 0, size_t height = 0) {
	unsigned int buf = 0;
	// store image
	if (!buffer->getIsIncomplete() && buffer->getImagePresent(part)) {
		rcg::Image image(buffer, part);
		if (fmt == rcg::ImgFmt::PNG) {
			buf = storeBufferPNG(image, image_buffer, yoffset, height);
		}
		else if (fmt == rcg::ImgFmt::JPEG) {
			buf = storeBufferJPEG(image, image_buffer, yoffset, height);
		}
		else {
			buf = storeBufferPNM(image, image_buffer, yoffset, height);
		}
	}
	else if (buffer->getIsIncomplete()) {
		std::cout << "storeBuffer(): Received incomplete buffer" << std::endl;
	}
	else if (!buffer->getImagePresent(part)) {
		std::cout << "storeBuffer(): Received buffer without image" << std::endl;
	}
	return buf;
}

int storeBufferPNM(const rcg::Image& image, unsigned char** image_buffer, size_t yoffset, size_t height) {
	int size = 0;
	size_t width = image.getWidth();
	size_t real_height = image.getHeight();
	if (height == 0) height = real_height;
	yoffset = (((yoffset) < (real_height)) ? (yoffset) : (real_height));
	height = (((height) < (real_height - yoffset)) ? (height) : (real_height - yoffset));
	const unsigned char* p = static_cast<const unsigned char*>(image.getPixels());
	unsigned char* buf = NULL;
	size_t px = image.getXPadding();
	uint64_t format = image.getPixelFormat();

	switch (format)
	{
	case Mono8:
	case Confidence8:
	case Error8:
	{
		std::string s = "P5\n" + std::to_string(width) + " " + std::to_string(height) + "\n" + std::to_string(255) + "\n";
		*image_buffer = (unsigned char*)malloc(s.length() + height * width);
		if (NULL == *image_buffer) {
			std::cerr << "malloc() memory allocation failed!" << std::endl;
			return MALLAC_ERROR;
		}
		size = int(s.length() + height * width);
		buf = *image_buffer;
		memcpy(buf, s.c_str(), s.length());
		buf += s.length();

		p += (width + px) * yoffset;
		for (size_t k = 0; k < height; k++) {
			for (size_t i = 0; i < width; i++) {
				*buf++ = *p++;
			}
			p += px;
		}
	}
	break;

	case Mono16:
	case Coord3D_C16:
	{
		std::string s = "P5\n" + std::to_string(width) + " " + std::to_string(height) + "\n" + std::to_string(65535) + "\n";
		*image_buffer = (unsigned char*)malloc(s.length() + 2 * height * width);
		if (NULL == *image_buffer) {
			std::cerr << "malloc() memory allocation failed!" << std::endl;
			return MALLAC_ERROR;
		}
		size = int(s.length() + 2 * height * width);
		buf = *image_buffer;
		memcpy(buf, s.c_str(), s.length());
		buf += s.length();

		// copy image data, pgm is always big endian
		p += (2 * width + px) * yoffset;
		if (image.isBigEndian()) {
			for (size_t k = 0; k < height; k++) {
				for (size_t i = 0; i < width; i++) {
					*buf++ = *p++;
					*buf++ = *p++;
				}
				p += px;
			}
		}
		else {
			for (size_t k = 0; k < height; k++) {
				for (size_t i = 0; i < width; i++) {
					*buf++ = p[1];
					*buf++ = p[0];
					p += 2;
				}
				p += px;
			}
		}
	}
	break;

	case YCbCr411_8:
	case YCbCr422_8:
	case YUV422_8:
	{
		std::string s = "P6\n" + std::to_string(width) + " " + std::to_string(height) + "\n" + std::to_string(255) + "\n";
		*image_buffer = (unsigned char*)malloc(s.length() + 3 * height * width);
		if (NULL == *image_buffer) {
			std::cerr << "malloc() memory allocation failed!" << std::endl;
			return MALLAC_ERROR;
		}
		size = int(s.length() + 3 * height * width);
		buf = *image_buffer;
		memcpy(buf, s.c_str(), s.length());
		buf += s.length();

		size_t pstep;
		if (format == YCbCr411_8) {
			pstep = (width >> 2) * 6 + px;
		}
		else {
			pstep = (width >> 2) * 8 + px;
		}

		p += pstep * yoffset;
		for (size_t k = 0; k < height; k++) {
			for (size_t i = 0; i < width; i += 4) {
				uint8_t rgb[12];

				if (format == YCbCr411_8) {
					rcg::convYCbCr411toQuadRGB(rgb, p, static_cast<int>(i));
				}
				else {
					rcg::convYCbCr422toQuadRGB(rgb, p, static_cast<int>(i));
				}

				for (int j = 0; j < 12; j++) {
					*buf++ = rgb[j];
				}
			}

			p += pstep;
		}

	}
	break;

	default:
	{
		std::unique_ptr<uint8_t[]> rgb_pixel(new uint8_t[3 * width * height]);

		if (format == RGB8) {
			p += (3 * width + px) * yoffset;
		}
		else {
			p += (width + px) * yoffset;
		}

		if (rcg::convertImage(rgb_pixel.get(), 0, p, format, width, height, px)) {
			p = rgb_pixel.get();
			std::string s = "P6\n" + std::to_string(width) + " " + std::to_string(height) + "\n" + std::to_string(255) + "\n";
			*image_buffer = (unsigned char*)malloc(s.length() + 3 * height * width);
			if (NULL == *image_buffer) {
				std::cerr << "malloc() memory allocation failed!" << std::endl;
				return MALLAC_ERROR;
			}
			size = int(s.length() + 3 * height * width);
			buf = *image_buffer;
			memcpy(buf, s.c_str(), s.length());
			buf += s.length();
			for (size_t k = 0; k < height; k++) {
				for (size_t i = 0; i < width; i++) {
					*buf++ = *p++;
					*buf++ = *p++;
					*buf++ = *p++;
				}
			}
		}
		else {
			throw IOException(std::string("storeImage(): Unsupported pixel format: ") +
				GetPixelFormatName(static_cast<PfncFormat>(image.getPixelFormat())));
		}
	}
	break;
	}

	return size;
}

int storeBufferPNG(const rcg::Image& image, unsigned char** image_buffer, size_t yoffset, size_t height) {
	unsigned int size = 0;
	size_t width = image.getWidth();
	size_t real_height = image.getHeight();
	if (height == 0) height = real_height;
	yoffset = (((yoffset) < (real_height)) ? (yoffset) : (real_height));
	height = (((height) < (real_height - yoffset)) ? (height) : (real_height - yoffset));
	const unsigned char* p = static_cast<const unsigned char*>(image.getPixels());
	size_t px = image.getXPadding();
	uint64_t format = image.getPixelFormat();

	switch (format)
	{
	case Mono8:
	case Confidence8:
	case Error8:
	{
		// open file and init
		std::string full_name = "test.png";
		FILE* out = fopen(full_name.c_str(), "wb+");
		if (!out) {
			throw new IOException("Cannot store file: " + full_name);
		}
		png_structp png = png_create_write_struct(PNG_LIBPNG_VER_STRING, 0, 0, 0);
		png_infop info = png_create_info_struct(png);
		unsigned char* png_s = (unsigned char*)png;
		setjmp(png_jmpbuf(png));

		// write header
		png_init_io(png, out);
		png_set_IHDR(png, info, width, height, 8, PNG_COLOR_TYPE_GRAY,
			PNG_INTERLACE_NONE, PNG_COMPRESSION_TYPE_DEFAULT,
			PNG_FILTER_TYPE_DEFAULT);
		png_write_info(png, info);

		// write image body
		p += (width + px) * yoffset;
		for (size_t k = 0; k < height; k++) {
			png_write_row(png, const_cast<png_bytep>(p));
			p += width + px;
		}

		// close file
		png_write_end(png, info);
		fseek(out, 0, SEEK_END);
		size = (unsigned int)ftell(out);
		*image_buffer = (unsigned char*)malloc(size * sizeof(char));
		if (NULL == *image_buffer) {
			std::cerr << "malloc() memory allocation failed!" << std::endl;
			fclose(out);
			png_destroy_write_struct(&png, &info);
			return MALLAC_ERROR;
		}
		rewind(out);
		fread(*image_buffer, 1, size, out);

		fclose(out);
		png_destroy_write_struct(&png, &info);
	}
	break;

	case Mono16:
	case Coord3D_C16:
	{
		std::string full_name = "test.png";
		FILE* out = fopen(full_name.c_str(), "wb+");
		if (!out) {
			throw new IOException("Cannot store file: " + full_name);
		}

		png_structp png = png_create_write_struct(PNG_LIBPNG_VER_STRING, 0, 0, 0);
		png_infop info = png_create_info_struct(png);
		setjmp(png_jmpbuf(png));

		png_init_io(png, out);
		png_set_IHDR(png, info, width, height, 16, PNG_COLOR_TYPE_GRAY,
			PNG_INTERLACE_NONE, PNG_COMPRESSION_TYPE_DEFAULT,
			PNG_FILTER_TYPE_DEFAULT);
		png_write_info(png, info);

		if (!image.isBigEndian()) {
			png_set_swap(png);
		}

		p += (2 * width + px) * yoffset;
		for (size_t k = 0; k < height; k++) {
			png_write_row(png, const_cast<png_bytep>(p));
			p += 2 * width + px;
		}

		png_write_end(png, info);
		fseek(out, 0, SEEK_END);
		size = (unsigned int)ftell(out);
		*image_buffer = (unsigned char*)malloc(size * sizeof(char));
		if (NULL == *image_buffer) {
			std::cerr << "malloc() memory allocation failed!" << std::endl;
			fclose(out);
			png_destroy_write_struct(&png, &info);
			return MALLAC_ERROR;
		}
		rewind(out);
		fread(*image_buffer, 1, size, out);

		fclose(out);
		png_destroy_write_struct(&png, &info);
	}
	break;

	case YCbCr411_8:
	case YCbCr422_8:
	case YUV422_8:
	{
		std::string full_name = "test.png";
		FILE* out = fopen(full_name.c_str(), "wb+");
		if (!out) {
			throw new IOException("Cannot store file: " + full_name);
		}

		png_structp png = png_create_write_struct(PNG_LIBPNG_VER_STRING, 0, 0, 0);
		png_infop info = png_create_info_struct(png);
		setjmp(png_jmpbuf(png));

		png_init_io(png, out);
		png_set_IHDR(png, info, width, height, 8, PNG_COLOR_TYPE_RGB,
			PNG_INTERLACE_NONE, PNG_COMPRESSION_TYPE_DEFAULT,
			PNG_FILTER_TYPE_DEFAULT);
		png_write_info(png, info);


		uint8_t* tmp = new uint8_t[3 * width];

		size_t pstep;
		if (format == YCbCr411_8) {
			pstep = (width >> 2) * 6 + px;
		}
		else {
			pstep = (width >> 2) * 8 + px;
		}

		p += pstep * yoffset;
		for (size_t k = 0; k < height; k++) {
			if (format == YCbCr411_8) {
				for (size_t i = 0; i < width; i += 4) {
					rcg::convYCbCr411toQuadRGB(tmp + 3 * i, p, static_cast<int>(i));
				}
			}
			else {
				for (size_t i = 0; i < width; i += 4) {
					rcg::convYCbCr422toQuadRGB(tmp + 3 * i, p, static_cast<int>(i));
				}
			}

			png_write_row(png, tmp);
			p += pstep;
		}

		// close file
		png_write_end(png, info);

		fseek(out, 0, SEEK_END);
		size = (unsigned int)ftell(out);
		*image_buffer = (unsigned char*)malloc(size * sizeof(char));
		if (NULL == *image_buffer) {
			std::cerr << "malloc() memory allocation failed!" << std::endl;
			fclose(out);
			png_destroy_write_struct(&png, &info);
			return MALLAC_ERROR;
		}
		rewind(out);
		fread(*image_buffer, 1, size, out);

		fclose(out);
		png_destroy_write_struct(&png, &info);
	}
	break;

	default:
	{
		std::unique_ptr<uint8_t[]> rgb_pixel(new uint8_t[3 * width * height]);

		if (format == RGB8) {
			p += (3 * width + px) * yoffset;
		}
		else {
			p += (width + px) * yoffset;
		}

		if (rcg::convertImage(rgb_pixel.get(), 0, p, format, width, height, px)) {
			p = rgb_pixel.get();
			std::string full_name = "test.png";
			FILE* out = fopen(full_name.c_str(), "wb+");
			if (!out) {
				throw new IOException("Cannot store file: " + full_name);
			}

			png_structp png = png_create_write_struct(PNG_LIBPNG_VER_STRING, 0, 0, 0);
			png_infop info = png_create_info_struct(png);
			setjmp(png_jmpbuf(png));

			png_init_io(png, out);
			png_set_IHDR(png, info, width, height, 8, PNG_COLOR_TYPE_RGB,
				PNG_INTERLACE_NONE, PNG_COMPRESSION_TYPE_DEFAULT,
				PNG_FILTER_TYPE_DEFAULT);
			png_write_info(png, info);

			for (size_t k = 0; k < height; k++) {
				png_write_row(png, p);
				p += 3 * width;
			}

			png_write_end(png, info);

			fseek(out, 0, SEEK_END);
			size = (unsigned int)ftell(out);
			*image_buffer = (unsigned char*)malloc(size * sizeof(char));
			if (NULL == *image_buffer) {
				std::cerr << "malloc() memory allocation failed!" << std::endl;
				fclose(out);
				png_destroy_write_struct(&png, &info);
				return MALLAC_ERROR;
			}
			rewind(out);
			fread(*image_buffer, 1, size, out);

			fclose(out);
			png_destroy_write_struct(&png, &info);
		}
		else {
			throw IOException(std::string("storeImage(): Unsupported pixel format: ") +
				GetPixelFormatName(static_cast<PfncFormat>(image.getPixelFormat())));
		}
	}
	break;
	}
	remove("test.png");
	return size;
}

int storeBufferJPEG(const rcg::Image& image, unsigned char** image_buffer, size_t yoffset, size_t height) {
	unsigned int size = 0;
	size_t width = image.getWidth();
	size_t real_height = image.getHeight();
	if (height == 0) height = real_height;
	yoffset = (((yoffset) < (real_height)) ? (yoffset) : (real_height));
	height = (((height) < (real_height - yoffset)) ? (height) : (real_height - yoffset));
	const unsigned char* p = static_cast<const unsigned char*>(image.getPixels());
	size_t px = image.getXPadding();
	uint64_t format = image.getPixelFormat();

	switch (format)
	{
	case Mono8:
	case Confidence8:
	case Error8:
	{
		//initialize JPEG compression objects, and specify the error handler
		jpeg_compress_struct jpeg{};
		jpeg_error_mgr jerr;
		jpeg.err = jpeg_std_error(&jerr);
		jpeg_create_compress(&jpeg);

		//specify the target file for JPEG storage
		std::string full_name = "test.jpeg";
		FILE* out = fopen(full_name.c_str(), "wb+");
		if (!out) {
			throw new IOException("Cannot store file: " + full_name);
		}
		jpeg_stdio_dest(&jpeg, out);

		/*
		Set compression parameters:
		image width, 
		image height, 
		number of color channels (gray image 1, color image 3), 
		color space (jcs_grayscale represents gray image, jcs_rgb represents color image), 
		compression quality
		*/
		jpeg.image_width = width;
		jpeg.image_height = height;
		//std::cout << jpeg.image_width << "  " << jpeg.image_width << std::endl;
		jpeg.input_components = 1;
		jpeg.in_color_space = JCS_GRAYSCALE;
		jpeg_set_defaults(&jpeg);
		jpeg_set_quality(&jpeg, 100, TRUE);

		// write image body
		JSAMPROW row_pointer;
		jpeg_start_compress(&jpeg, TRUE);
		p += (width + px) * yoffset;
		for (size_t k = 0; k < height; k++) {
			row_pointer = (unsigned char*)p;
			jpeg_write_scanlines(&jpeg, &row_pointer, 1);
			p += width + px;
		}

		// close file
		fseek(out, 0, SEEK_END);
		size = (unsigned int)ftell(out);
		*image_buffer = (unsigned char*)malloc(size * sizeof(char));
		if (NULL == *image_buffer) {
			std::cerr << "malloc() memory allocation failed!" << std::endl;
			fclose(out);
			jpeg_destroy_compress(&jpeg);
			return MALLAC_ERROR;
		}
		rewind(out);
		fread(*image_buffer, 1, size, out);
		fclose(out);
		jpeg_destroy_compress(&jpeg);
	}
	break;

	case Mono16:
	case Coord3D_C16:
	{
		jpeg_compress_struct jpeg{};
		jpeg_error_mgr jerr;
		jpeg.err = jpeg_std_error(&jerr);
		jpeg_create_compress(&jpeg);

		std::string full_name = "test.jpeg";
		FILE* out = fopen(full_name.c_str(), "wb+");
		if (!out) {
			throw new IOException("Cannot store file: " + full_name);
		}
		jpeg_stdio_dest(&jpeg, out);

		jpeg.image_width = width;
		jpeg.image_height = height;
		jpeg.input_components = 2;
		jpeg.in_color_space = JCS_GRAYSCALE;
		jpeg_set_defaults(&jpeg);
		jpeg_set_quality(&jpeg, 100, TRUE);

		JSAMPROW row_pointer;
		jpeg_start_compress(&jpeg, TRUE);
		p += (2 * width + px) * yoffset;
		for (size_t k = 0; k < height; k++) {
			row_pointer = (unsigned char*)p;
			jpeg_write_scanlines(&jpeg, &row_pointer, 1);
			p += 2 * width + px;
		}

		fseek(out, 0, SEEK_END);
		size = (unsigned int)ftell(out);
		*image_buffer = (unsigned char*)malloc(size * sizeof(char));
		if (NULL == *image_buffer) {
			std::cerr << "malloc() memory allocation failed!" << std::endl;
			fclose(out);
			jpeg_destroy_compress(&jpeg);
			return MALLAC_ERROR;
		}
		rewind(out);
		fread(*image_buffer, 1, size, out);
		fclose(out);
		jpeg_destroy_compress(&jpeg);
	}
	break;

	case YCbCr411_8:
	case YCbCr422_8:
	case YUV422_8:
	{
		jpeg_compress_struct jpeg{};
		jpeg_error_mgr jerr;
		jpeg.err = jpeg_std_error(&jerr);
		jpeg_create_compress(&jpeg);

		std::string full_name = "test.jpeg";
		FILE* out = fopen(full_name.c_str(), "wb+");
		if (!out) {
			throw new IOException("Cannot store file: " + full_name);
		}
		jpeg_stdio_dest(&jpeg, out);

		jpeg.image_width = width;
		jpeg.image_height = height;
		jpeg.input_components = 3;
		jpeg.in_color_space = JCS_RGB;
		jpeg_set_defaults(&jpeg);
		jpeg_set_quality(&jpeg, 100, TRUE);

		uint8_t* tmp = new uint8_t[3 * width];
		size_t pstep;
		if (format == YCbCr411_8) {
			pstep = (width >> 2) * 6 + px;
		}
		else {
			pstep = (width >> 2) * 8 + px;
		}

		JSAMPROW row_pointer;
		jpeg_start_compress(&jpeg, TRUE);
		p += pstep * yoffset;
		for (size_t k = 0; k < height; k++) {
			if (format == YCbCr411_8) {
				for (size_t i = 0; i < width; i += 4) {
					rcg::convYCbCr411toQuadRGB(tmp + 3 * i, p, static_cast<int>(i));
				}
			}
			else {
				for (size_t i = 0; i < width; i += 4) {
					rcg::convYCbCr422toQuadRGB(tmp + 3 * i, p, static_cast<int>(i));
				}
			}
			row_pointer = (unsigned char*)p;
			jpeg_write_scanlines(&jpeg, &row_pointer, 1);
			p += pstep;
		}

		fseek(out, 0, SEEK_END);
		size = (unsigned int)ftell(out);
		*image_buffer = (unsigned char*)malloc(size * sizeof(char));
		if (NULL == *image_buffer) {
			std::cerr << "malloc() memory allocation failed!" << std::endl;
			fclose(out);
			jpeg_destroy_compress(&jpeg);
			return MALLAC_ERROR;
		}
		rewind(out);
		fread(*image_buffer, 1, size, out);
		fclose(out);
		jpeg_destroy_compress(&jpeg);
	}
	break;

	default:
	{
		std::unique_ptr<uint8_t[]> rgb_pixel(new uint8_t[3 * width * height]);
		if (format == RGB8) {
			p += (3 * width + px) * yoffset;
		}
		else {
			p += (width + px) * yoffset;
		}
		if (rcg::convertImage(rgb_pixel.get(), 0, p, format, width, height, px)) {
			p = rgb_pixel.get();
			jpeg_compress_struct jpeg{};
			jpeg_error_mgr jerr;
			jpeg.err = jpeg_std_error(&jerr);
			jpeg_create_compress(&jpeg);

			std::string full_name = "test.jpeg";
			FILE* out = fopen(full_name.c_str(), "wb+");
			if (!out) {
				throw new IOException("Cannot store file: " + full_name);
			}
			jpeg_stdio_dest(&jpeg, out);

			jpeg.image_width = width;
			jpeg.image_height = height;
			jpeg.input_components = 3;
			jpeg.in_color_space = JCS_RGB;
			jpeg_set_defaults(&jpeg);
			jpeg_set_quality(&jpeg, 100, TRUE);

			JSAMPROW row_pointer;
			jpeg_start_compress(&jpeg, TRUE);
			for (size_t k = 0; k < height; k++) {
				row_pointer = (unsigned char*)p;
				jpeg_write_scanlines(&jpeg, &row_pointer, 1);
				p += 3 * width;
			}

			fseek(out, 0, SEEK_END);
			size = (unsigned int)ftell(out);
			*image_buffer = (unsigned char*)malloc(size * sizeof(char));
			if (NULL == *image_buffer) {
				std::cerr << "malloc() memory allocation failed!" << std::endl;
				fclose(out);
				jpeg_destroy_compress(&jpeg);
				return MALLAC_ERROR;
			}
			rewind(out);
			fread(*image_buffer, 1, size, out);
			fclose(out);
			jpeg_destroy_compress(&jpeg);
		}
		else {
			throw IOException(std::string("storeImage(): Unsupported pixel format: ") +
				GetPixelFormatName(static_cast<PfncFormat>(image.getPixelFormat())));
		}
	}
	break;
	}
	remove("test.jpeg");
	return size;
}

extern "C" int get_image(MyDevice myDevice, const char* imgfmt, unsigned char** image_buffer, int* size, char** err) {
	std::string e = "";
	int ret = SUCCESS;
	try {
		*err = (char*)malloc(e.length() + 1);
		strcpy(*err, e.c_str());
		if (NULL == *err) {
			std::cerr << "malloc() memory allocation failed!" << std::endl;
			return MALLAC_ERROR;
		}
		rcg::ImgFmt fmt = rcg::ImgFmt::PNM;
		//std::cout << imgfmt << std::endl;
		if (!strncmp(imgfmt, "pnm", 3)) {
			fmt = rcg::ImgFmt::PNM;
		}
		else if (!strncmp(imgfmt, "png", 3)) {
			fmt = rcg::ImgFmt::PNG;
		}
		else if (!strncmp(imgfmt, "jpeg", 4)) {
			fmt = rcg::ImgFmt::JPEG;
		}

		// open stream and get 1 image
		std::vector<std::shared_ptr<rcg::Stream> > stream = myDevice.dev->getStreams();
		if (stream.size() > 0) {
			stream[0]->open();
			stream[0]->attachBuffers(true);
			stream[0]->startStreaming();

			int buffers_received = 0;
			int buffers_incomplete = 0;

			// grab next image with timeout of 3 seconds
			int retry = 10;
			while (retry > 0) {
				const rcg::Buffer* buffer = stream[0]->grab(3000);

				if (buffer != 0) {
					buffers_received++;
					if (!buffer->getIsIncomplete()) {
						// store images
						*size = storeBuffer(fmt, buffer, image_buffer, 0);

						if (*size == 0) {
							buffers_incomplete++;
							ret = 1;
						}
						else if (*size == -1) {
							stream[0]->stopStreaming();
							stream[0]->close();
							return MALLAC_ERROR;
						}
						retry = 0;
					}
					else {
						std::cout << "Incomplete buffer received" << std::endl;
						buffers_incomplete++;
					}
				}
				else {
					std::cout << "Cannot grab images" << std::endl;
					break;
				}
				retry--;
			}

			stream[0]->stopStreaming();
			stream[0]->close();

			std::cout << std::endl;
			std::cout << "Received buffers:   " << buffers_received << std::endl;
			std::cout << "Incomplete buffers: " << buffers_incomplete << std::endl;

			// return error code if no images could be received
			if (buffers_incomplete == buffers_received) {
				ret = NO_IMAGES;
				std::string e = "No images could be received!";
				*err = (char*)malloc(e.length() + 1);
				if (NULL == *err) {
					std::cerr << "malloc() memory allocation failed!" << std::endl;
					return MALLAC_ERROR;
				}
				strcpy(*err, e.c_str());
			}
		}
		else {
			ret = NO_STREAMS;
			//std::cout << "No streams available" << std::endl;
			std::string e = "No streams available!";
			*err = (char*)malloc(e.length() + 1);
			if (NULL == *err) {
				std::cerr << "malloc() memory allocation failed!" << std::endl;
				return MALLAC_ERROR;
			}
			strcpy(*err, e.c_str());
		}
	}
	catch (const IOException& ex) {
		ret = IOEXCEPTION;
		//std::cout << "Exception: " << ex.what() << std::endl;
		std::string e = "IOException: " + (std::string)ex.what();
		*err = (char*)malloc(e.length() + 1);
		if (NULL == *err) {
			std::cerr << "malloc() memory allocation failed!" << std::endl;
			return MALLAC_ERROR;
		}
		strcpy(*err, e.c_str());
	}
	catch (const GENICAM_NAMESPACE::GenericException& ex) {
		ret = GEN_EXCEPTION;
		//std::cout << "Exception: " << ex.what() << std::endl;
		std::string e = "GenericException: " + (std::string)ex.what();
		*err = (char*)malloc(e.length() + 1);
		if (NULL == *err) {
			std::cerr << "malloc() memory allocation failed!" << std::endl;
			return MALLAC_ERROR;
		}
		strcpy(*err, e.c_str());
	}
	catch (const std::exception& ex) {
		ret = EXCEPTION;
		//std::cout << "Exception: " << ex.what() << std::endl;
		std::string e = "Exception: " + (std::string)ex.what();
		*err = (char*)malloc(e.length() + 1);
		if (NULL == *err) {
			std::cerr << "malloc() memory allocation failed!" << std::endl;
			return MALLAC_ERROR;
		}
		strcpy(*err, e.c_str());
	}
	return ret;

}

void free_image(unsigned char** image_buffer) {
	free(*image_buffer);
}
