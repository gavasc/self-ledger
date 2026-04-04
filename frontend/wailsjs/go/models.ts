export namespace main {
	
	export class Account {
	    id?: number;
	    name: string;
	    initial_balance: number;
	
	    static createFrom(source: any = {}) {
	        return new Account(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.initial_balance = source["initial_balance"];
	    }
	}
	export class AccountBalance {
	    id: number;
	    name: string;
	    balance: number;
	
	    static createFrom(source: any = {}) {
	        return new AccountBalance(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.balance = source["balance"];
	    }
	}
	export class Installment {
	    id?: number;
	    desc: string;
	    cat: string;
	    total_val: number;
	    n_installments: number;
	    start_date: string;
	    account_id?: number;
	    paid_count?: number;
	    monthly_val?: number;
	
	    static createFrom(source: any = {}) {
	        return new Installment(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.desc = source["desc"];
	        this.cat = source["cat"];
	        this.total_val = source["total_val"];
	        this.n_installments = source["n_installments"];
	        this.start_date = source["start_date"];
	        this.account_id = source["account_id"];
	        this.paid_count = source["paid_count"];
	        this.monthly_val = source["monthly_val"];
	    }
	}
	export class Transaction {
	    id?: number;
	    type: string;
	    desc: string;
	    cat: string;
	    val: number;
	    date: string;
	    account_id?: number;
	    installment_id?: number;
	    installment_index?: number;
	
	    static createFrom(source: any = {}) {
	        return new Transaction(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.type = source["type"];
	        this.desc = source["desc"];
	        this.cat = source["cat"];
	        this.val = source["val"];
	        this.date = source["date"];
	        this.account_id = source["account_id"];
	        this.installment_id = source["installment_id"];
	        this.installment_index = source["installment_index"];
	    }
	}
	export class Transfer {
	    id?: number;
	    from_account_id: number;
	    to_account_id: number;
	    from_account_name?: string;
	    to_account_name?: string;
	    amount: number;
	    date: string;
	    desc: string;
	
	    static createFrom(source: any = {}) {
	        return new Transfer(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.from_account_id = source["from_account_id"];
	        this.to_account_id = source["to_account_id"];
	        this.from_account_name = source["from_account_name"];
	        this.to_account_name = source["to_account_name"];
	        this.amount = source["amount"];
	        this.date = source["date"];
	        this.desc = source["desc"];
	    }
	}

}

