Courses Marketplace

This repository contains a server written in Go for a courses marketplace. The main Go service is built with the Gin framework, following the Clean architecture design pattern. The backend includes stateful authentication, allowing users to register and access their personal accounts. The personal account provides access to purchased courses and offers the option to acquire additional video lessons. Video playback is implemented using HTTP Live Streaming, and a custom video player is developed, taking inspiration from the Plyr player.

The repository also features an admin panel with various functionalities. Through the admin panel, administrators can upload courses. Upon upload, videos are asynchronously processed using FFmpeg to create three different quality variants. The videos are then asynchronously uploaded to Yandex Cloud's S3 storage. Additionally, the admin panel allows for user management, including the ability to add, delete, and modify user accounts, among other features.
Features

    User registration and stateful authentication
    Personal account for accessing purchased courses and acquiring new video lessons
    HTTP Live Streaming for video playback
    Custom video player inspired by Plyr
    Admin panel for course management and user administration
    Asynchronous video processing using FFmpeg
    Asynchronous video upload to Yandex Cloud's S3 storage

Technologies Used

    Go
    Gin framework
    Clean architecture
    HTTP Live Streaming (HLS)
    FFmpeg
    Yandex Cloud S3 storage

Contributions

Contributions to the Courses Marketplace project are welcome! If you have any ideas, suggestions, or improvements, feel free to open an issue or submit a pull request.

Please ensure that your contributions align with the coding style, guidelines, and best practices followed in the project.
