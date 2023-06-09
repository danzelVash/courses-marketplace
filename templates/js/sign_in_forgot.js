let first_email = ""

function get_code() {
    let email = document.querySelector('#email').value;
    first_email = email
    fetch("http://localhost/auth/sign-in/forgot-pass/send-code/", {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify({"email": email})
    })
        .then(res => {
            let response = JSON.stringify(res.body);
            if (res.status === 500) {
                console.log("happend something bad, try again")
            } else if (res.status === 200 && response["exist"] === true) {
                console.log("email is exist and code was sent")
            } else if (res.status === 400 && response["exist"] === false) {
                console.log("email is not registered in site")
            }
        })
}

function check_code() {
    let code = document.querySelector('#code').value;

    fetch("http://localhost/auth/sign-in/forgot-pass/", {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify({"email": first_email, "code": code})
    })
        .then(res => {
            if (res.redirected) {
                document.location = res.url
            }
        })
}