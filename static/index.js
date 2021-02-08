function getEntries(url, method, data) {
    return axios({
        method: method,
        url: url,
        data: data,
        xsrfCookieName: 'csrftoken',
        xsrfHeaderName: 'X-CSRFToken',
        headers: {
            'X-Requested-With': 'XMLHttpRequest'
        }
    })
}

var app = new Vue({
    el: '#app',
    data: {
        keyword: '',
        entries: [
            { code: 'LOADING', name: 'LOADING', open: 'LOADING', high: 'LOADING', low: 'LOADING', close:  'LOADING'},
        ]
    },
    created() {
        let r = getEntries('', 'get', '').then((res) => {
            this.entries = res.data.entries
        });
    },
    computed: {
        filteredEntries() {
            return this.entries.filter(entry => {
                return entry.name.toLowerCase().includes(this.keyword.toLowerCase())
            })
        }
    },
    methods: {
        downloadCsv() {
            const items = this.filteredEntries;
            const replacer = (_, value) => value === null ? '' : value;
            const header = Object.keys(items[0]);
            const csv = [
                header.join(','),
                ...items.map(row => header.map(fieldName => JSON.stringify(row[fieldName], replacer)).join(', '))
            ].join('\r\n').replace(/['"]+/g, '');

            let element = document.createElement('a');
            element.style.display = 'none';
            element.setAttribute('href', 'data:text/plain;charset=utf-8,' + encodeURIComponent(csv));
            element.setAttribute('download', 'output.csv');
            
            document.body.appendChild(element);
            element.click();
            document.body.removeChild(element);
        }
    }
})