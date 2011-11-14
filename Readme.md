Early progress on an app whose sole purpose is to "get me this file from S3 as fast as possible".

## Usage:

    $ ./s3me
    Usage of ./s3me:
      -c=1: Number of connections
      -o="./output": Output path
      -s=0: Number of segments
      -u="": -u https://bucket.s3.amazonaws.com/key

    $ ./s3me -u https://bucket-name.s3.amazonaws.com/filename.mp4 -c 4 -s 16 -o ~/Desktop/output.mp4

