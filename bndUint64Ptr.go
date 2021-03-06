// Copyright 2014 Rana Ian. All rights reserved.
// Use of this source code is governed by The MIT License
// found in the accompanying LICENSE file.

package ora

/*
#include <stdlib.h>
#include <oci.h>
#include "version.h"
*/
import "C"
import "unsafe"

type bndUint64Ptr struct {
	stmt      *Stmt
	ocibnd    *C.OCIBind
	ociNumber [1]C.OCINumber
	value     *uint64
	nullp
}

func (bnd *bndUint64Ptr) bind(value *uint64, position namedPos, stmt *Stmt) error {
	//bnd.stmt.logF(_drv.Cfg().Log.Stmt.Bind, "Uint64Ptr.bind(%d) value=%#v => number=%#v", position, value, bnd.ociNumber[0])
	bnd.stmt = stmt
	bnd.value = value
	bnd.nullp.Set(value == nil)
	if value != nil {
		if err := bnd.stmt.ses.srv.env.OCINumberFromInt(&bnd.ociNumber[0], intSixtyFour(*value), byteWidth64); err != nil {
			return err
		}
		bnd.stmt.logF(_drv.Cfg().Log.Stmt.Bind,
			"Uint64Ptr.bind(%s) value=%#v => number=%#v", position, value, bnd.ociNumber[0])
	}
	ph, phLen, phFree := position.CString()
	if ph != nil {
		defer phFree()
	}
	r := C.bindByNameOrPos(
		bnd.stmt.ocistmt, //OCIStmt      *stmtp,
		&bnd.ocibnd,
		bnd.stmt.ses.srv.env.ocierr, //OCIError     *errhp,
		C.ub4(position.Ordinal),     //ub4          position,
		ph,
		phLen,
		unsafe.Pointer(&bnd.ociNumber[0]),   //void         *valuep,
		C.LENGTH_TYPE(C.sizeof_OCINumber),   //sb8          value_sz,
		C.SQLT_VNU,                          //ub2          dty,
		unsafe.Pointer(bnd.nullp.Pointer()), //void         *indp,
		nil,           //ub2          *alenp,
		nil,           //ub2          *rcodep,
		0,             //ub4          maxarr_len,
		nil,           //ub4          *curelep,
		C.OCI_DEFAULT) //ub4          mode );
	if r == C.OCI_ERROR {
		return bnd.stmt.ses.srv.env.ociError()
	}
	return nil
}

func (bnd *bndUint64Ptr) setPtr() error {
	if bnd.nullp.IsNull() {
		return nil
	}
	i, err := bnd.stmt.ses.srv.env.OCINumberToInt(&bnd.ociNumber[0], byteWidth64)
	*bnd.value = uint64(i)
	return err
}

func (bnd *bndUint64Ptr) close() (err error) {
	defer func() {
		if value := recover(); value != nil {
			err = errR(value)
		}
	}()

	stmt := bnd.stmt
	bnd.stmt = nil
	bnd.ocibnd = nil
	bnd.value = nil
	bnd.nullp.Free()
	stmt.putBnd(bndIdxUint64Ptr, bnd)
	return nil
}
