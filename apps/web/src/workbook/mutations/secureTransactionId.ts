import { clientTxnID } from "../../services/browserApi";

/**
 * Private assembly dependency for logical mutation identity.
 *
 * Presentation owners receive semantic mutation commands, never this port.
 */
export interface SecureTransactionIdPort {
  create(prefix: string): string;
}

export function createBrowserSecureTransactionIdPort(): SecureTransactionIdPort {
  return {
    create: clientTxnID,
  };
}
