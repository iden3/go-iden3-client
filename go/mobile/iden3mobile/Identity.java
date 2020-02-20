// Code generated by gobind. DO NOT EDIT.

// Java class iden3mobile.Identity is a proxy for talking to a Go program.
//
//   autogenerated by gobind -lang=java github.com/arnaubennassar/go-playground/mobile/mobile
package iden3mobile;

import go.Seq;

public final class Identity implements Seq.Proxy {
	static { Iden3mobile.touch(); }
	
	private final int refnum;
	
	@Override public final int incRefnum() {
	      Seq.incGoRef(refnum, this);
	      return refnum;
	}
	
	/**
	 * NewIdentity creates a new identity
	this funciton is mapped as a constructor in Java
	 */
	public Identity(String storePath, String pass, BytesArray extraGenesisClaims, Event e) {
		this.refnum = __NewIdentity(storePath, pass, extraGenesisClaims, e);
		Seq.trackGoRef(refnum, this);
	}
	
	private static native int __NewIdentity(String storePath, String pass, BytesArray extraGenesisClaims, Event e);
	
	/**
	 * NewIdentityLoad loads an already created identity
	this funciton is mapped as a constructor in Java
	 */
	public Identity(String storePath, Event e) {
		this.refnum = __NewIdentityLoad(storePath, e);
		Seq.trackGoRef(refnum, this);
	}
	
	private static native int __NewIdentityLoad(String storePath, Event e);
	
	Identity(int refnum) { this.refnum = refnum; Seq.trackGoRef(refnum, this); }
	
	public final native TicketsMap getTickets();
	public final native void setTickets(TicketsMap v);
	
	public native byte[] getReceivedClaim(long pos) throws Exception;
	public native long getReceivedClaimsLen();
	public native void proveClaim(String endpoint, long credIndex, Callback c);
	public native Ticket requestClaim(String endpoint, String data);
	@Override public boolean equals(Object o) {
		if (o == null || !(o instanceof Identity)) {
		    return false;
		}
		Identity that = (Identity)o;
		TicketsMap thisTickets = getTickets();
		TicketsMap thatTickets = that.getTickets();
		if (thisTickets == null) {
			if (thatTickets != null) {
			    return false;
			}
		} else if (!thisTickets.equals(thatTickets)) {
		    return false;
		}
		return true;
	}
	
	@Override public int hashCode() {
	    return java.util.Arrays.hashCode(new Object[] {getTickets()});
	}
	
	@Override public String toString() {
		StringBuilder b = new StringBuilder();
		b.append("Identity").append("{");
		b.append("Tickets:").append(getTickets()).append(",");
		return b.append("}").toString();
	}
}
