export const name = 'util';

export const BlankImage = "data:image/gif;base64,R0lGODlhAQABAAD/ACwAAAAAAQABAAACADs=";

export const SmallMediaLimit = 820;

export class Images {
    // Returns a promise which returns when an image is loaded
    static WaitForLoad(img) {
        return new Promise((resolve, reject) => {
            img.onload = function () {
                resolve()
            }
            img.onerror = reject
        })
    }

    static async LoadImages(images, total, urlFunc) {
        let promises = []
        for (let i = 0; i < total; i++) {
            let img = new Image()
            promises.push(Images.WaitForLoad(img))
            img.src = urlFunc(i)
            images.push(img)
        }
        return Promise.all(promises)
            .catch(error => { console.error("failed to load images:", error) }
        )
    }

    static BlankImageArray(length) {
        return Array.from({length:5}).map(x => {
            let img = new Image();
            img.src = BlankImage
            return img
        } )
    }
}

export class Compare {
    static EntryTitle(a, b) {
        return Compare.Strings(a.title, b.title)
    }

    static EntryOrder(a, b) {
        return Compare.Numbers(a.order, b.order)
    }

    static Users(a, b) {
       return Compare.Strings(a.name, b.name)
    }

    static Numbers(a, b) {
        return a - b
    }

    static Strings(a, b) {
        let x = a.toLowerCase()
        let y = b.toLowerCase()

        if (x < y) {
            return -1
        } else if (x > y) {
            return 1
        }
        return 0
    }
}

export class Fmt {
    static SeriesPercent(sp) {
        if (sp === undefined) {
            return "N/A"
        }

        let current = 0
        let total = 0
        for (let i in sp.tracker) {
            // .ensureSuccess function gets registered to
            // all objects and appears if we loop over an
            // so this avoids that
            if (sp.tracker.hasOwnProperty(i)) {
                let p = sp.tracker[i]
                current += p.current
                total += p.total
            }
        }

        return Fmt.Percent(current / total)
    }

    // Percent is supposed to be in the range [0, 1]
    static Percent(p) {
        if (p === undefined) {
            return "N/A"
        }
        return (p * 100).toFixed(2) + "%"
    }
}

export class Ensure {
    static Array(i) {
        if (i === undefined || i === null) {
            return []
        }
        return i
    }

    static Object(i) {
        if (i === undefined || i === null) {
            return {}
        }
        return i
    }
}

export class Search {
    static Match(search, text) {
        search = search.toUpperCase();
        text = text.toUpperCase();

        var j = -1; // remembers position of last found character

        // consider each search character one at a time
        for (var i = 0; i < search.length; i++) {
            var l = search[i];
            if (l == ' ') continue;     // ignore spaces

            j = text.indexOf(l, j+1);     // search for character & update position
            if (j == -1) return false;  // if it's not found, exclude this item
        }
        return true;
    }
}