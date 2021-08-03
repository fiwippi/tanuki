export const name = 'util';

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

// TODO use x-spread or something to create a reproducible library view
// TODO better search function that doesn't use startswith