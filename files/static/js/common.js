// Returns a promise which returns when an image is loaded
function waitUntilLoad(img) {
    return new Promise((resolve, reject) => {
        img.onload = function () {
            resolve()
        }
        img.onerror = reject
    })
}

// Used for comparing entries based on their title
function entryCompare(a, b) {
    return stringCompare(a.title, b.title)
}

function stringCompare(a, b) {
    let x = a.toLowerCase()
    let y = b.toLowerCase()

    if (x < y) {
        return -1
    } else if (x > y) {
        return 1
    }
    return 0
}

//
function stringPercent(p) {
    return percent(p).toString() + "%"
}

function percent(p) {
    return Math.round(p*100)/100
}

// Python like .format()
String.prototype.format = function() {
    let a = this;
    for (let k in arguments) {
        a = a.replace("{" + k + "}", arguments[k])
    }
    // if (a.startsWith("From")) {
    //     console.log(a, arguments)
    // }

    return a
}