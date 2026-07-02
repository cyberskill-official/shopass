export interface SourceFile {
    path: string;
    lines: string[];
}

export interface Violation {
    file?: string;
    line?: number;
    rule?: string;
    policyRef: string;
    description?: string;
}

export interface Manifest {
    permissions?: string[];
    host_permissions?: string[];
}

export function viol(scope: string, rule: string, policyRef: string): Violation {
    return {
        file: scope,
        rule,
        policyRef
    };
}
