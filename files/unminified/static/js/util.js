export const name = 'util';

export const BlankImage = "data:image/gif;base64,R0lGODlhAQABAAD/ACwAAAAAAQABAAACADs=";

// Navbar doesn't use this limit in order not to load Util to be more efficient
// so if changing this limit, Navbar needs to be updated manually
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

    static async LoadImages(images, total, urlFunc, replace) {
        let promises = []
        for (let i = 0; i < total; i++) {
            let img = new Image()
            promises.push(Images.WaitForLoad(img))
            img.src = urlFunc(i)
            replace ? images[i] = img : images.push(img)
        }
        return Promise.all(promises)
            .catch(error => { console.error("failed to load images:", error) }
        )
    }

    static BlankImageArray(length) {
        return Array.from({length: length}).map(x => {
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
    static SeriesPercent(sp, total_pages) {
        if (sp === undefined || sp === null) {
            return undefinedPercent
        }

        let current = 0
        for (let i in sp.tracker) {
            // .ensureSuccess function gets registered to
            // all objects and appears if we loop over an
            // so this avoids that
            if (sp.tracker.hasOwnProperty(i)) {
                let p = sp.tracker[i]
                if (p !== null && p !== undefined) {
                    current += p.current
                }
            }
        }

        let percent = current / total_pages
        if (Number.isNaN(percent)) {
            return undefinedPercent
        }
        return Fmt.Percent(percent)
    }

    // Percent is supposed to be in the range [0, 1]
    static EntryPercent(p) {
        if (p === undefined || p === null) {
            return undefinedPercent
        }
        let percent = p.current / p.total
        if (Number.isNaN(percent)) {
            return undefinedPercent
        }
        return Fmt.Percent(percent)
    }

    // Percent is supposed to be in the range [0, 1]
    static Percent(p) {
        if (p === undefined || p === null) {
            return undefinedPercent
        }
        return (p * 100).toFixed(2) + "%"
    }

    static RFCDate(date) {
        let d = new Date(date)

        let month = '' + (d.getMonth() + 1),
            day = '' + d.getDate(),
            year = d.getFullYear();

        if (month.length < 2)
            month = '0' + month;
        if (day.length < 2)
            day = '0' + day;

        return [year, month, day].join('-');
    }
}

const undefinedPercent = Fmt.Percent(0)

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

    static String(i) {
        if (i === undefined || i === null) {
            return ""
        }
        return i
    }

    static ScanlationGroup(r) {
        let g = r.relationships.find(o => { return o.type === "scanlation_group"})
        if (g === undefined)
            return ""
        return g.attributes.name
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
            if (l == ' ') continue;    // ignore spaces

            j = text.indexOf(l, j+1);  // search for character & update position
            if (j == -1) return false; // if it's not found, exclude this item
        }
        return true;
    }
}