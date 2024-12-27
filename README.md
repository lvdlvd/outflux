# outflux

Outflux is a tool to extract rows and columns from a influxdb line protocol formatted file

	https://docs.influxdata.com/influxdb/v1/write_protocols/line_protocol_tutorial/

Usage

```
	outflux [-csv|-tsv] [-m measurement] [tag=[val]]... [tag_or_fld [tag_or_fld]...]
```

Where

    tag=val   only process records that contain this tag with this value.
    tag=      only process records that contain this tag, irrespective of value
    tag~pfx   only process records that contain this tag with a value that has this prefix

    tag_or_fld  output this tag or field value as a column.

    only records that have all the specified columns will be output.

    if no tag_or_fld is specified, the program will output a summary over
    all records that match the filter.  The summary contains a histogram
    of all tags and their values, as well as the measurements and the record types,
    as defined by the set of fields they contain.

    No more than 19 values of each tag are shown

## Examples:

### Produce a summary over all records that have a tag 'trid' starting with ~PULSE

```
Luuk-van-Dijks-MacBook-Pro:outflux lvd$ gzcat ../dees/data/20240807111458/data.log.gz  | ./outflux trid~PULSE
outflux:processed 138231 of 3677327 records.
MEASUREMENTS:
	 138231 debug
RECORD TYPES:
	 133848 count,d,stamp,us
	   4383 hz
port:
	   4383 0
	 133848 4
session:
	 138231 raspberrypi
srcid:
	   4383 0
	 133848 23
start:
	 138231 20240807111459.330071510
trid:
	  69115 PULSEFALL
	  69116 PULSERAISE
outflux:Processed 967617 records in 2.750044167s
```

### Print a histogram of all trid values for fields that have a srcid tag:

```
Luuk-van-Dijks-MacBook-Pro:outflux lvd$ gzcat ../dees/data/20240807111458/data.log.gz  | ./outflux srcid= trid | sort | uniq -c
outflux:skipping record 3677328:tag at line 3677332:36: expected field key but none found
outflux:processed 2839762 of 3677327 records.
outflux:Processed 2839762 records in 2.883718084s
   1 #trid
3917 ACCELTEMP
444319 ACCEL_3G
19779 ANALOGIN
443049 BAROP
443048 BAROT
4391 DEVID
216890 DTS
440542 GYRO_250DEG_S
452251 MAGN
69115 PULSEFALL
69116 PULSERAISE
4389 STMTEMP
4388 SWREV
220178 TS
4390 VDD
```

### Print us, x, y, z fields of all ACCEL_ records from port 4

```
Luuk-van-Dijks-MacBook-Pro:outflux lvd$ go build &&  gzcat ../dees/data/20240807111458/data.log.gz  | ./outflux port=2 trid~ACCEL_ us x y z trid 
#us x y z trid
0 -1.0073547 0.07662964 -0.0026550293 ACCEL_3G
5802951255 -1.0081787 0.076812744 -0.0024719238 ACCEL_3G
5802956216 -1.0086365 0.0770874 -0.0018310547 ACCEL_3G
5802961180 -1.008728 0.07745361 -0.001373291 ACCEL_3G
5802966142 -1.0082703 0.07827759 -0.00036621094 ACCEL_3G
5802971101 -1.0075378 0.07873535 -0.0002746582 ACCEL_3G
5802976061 -1.006897 0.07937622 -0.00064086914 ACCEL_3G
5802981023 -1.0055237 0.08029175 -0.0011901855 ACCEL_3G
5802985985 -1.0045166 0.08102417 -0.002105713 ACCEL_3G
....
```


### Produce a summary over all records that have a tag 'trid' starting with ~PULSE

```
Luuk-van-Dijks-MacBook-Pro:outflux lvd$ go build &&  gzcat ../dees/data/20240807111458/data.log.gz  | ./outflux trid~PULSE us count | head
outflux:processed 138231 of 3677327 records.
MEASUREMENTS:
	 138231 debug
RECORD TYPES:
	 133848 count,d,stamp,us
	   4383 hz
port:
	   4383 0
	 133848 4
session:
	 138231 raspberrypi
srcid:
	   4383 0
	 133848 23
start:
	 138231 20240807111459.330071510
trid:
	  69115 PULSEFALL
	  69116 PULSERAISE
outflux:Processed 967617 records in 2.750044167s
```


### Produce a summary over all records

```
Luuk-van-Dijks-MacBook-Pro:outflux lvd$  gzcat ../dees/data/20240807111458/data.log.gz  | ./outflux
 gzcat ../dees/data/20240807111458/data.log.gz  | ./outflux
2024/12/27 01:18:20 skipping record 3677328:tag at line 3677332:36: expected field key but none found
2024/12/27 01:18:20 output 3677328 of 3677327 records.
MEASUREMENTS:
	3677328 debug
RECORD TYPES:
	  17587 V,d,stamp,us
	   1097 abrt,badmsg,bufsmall,crc,inv,noframe
	   5348 alt,alta,cr,d,diffAge,itow,lat,lata,lon,lona,numsv,sbgts,status,stid,type,und,used
	 106176 alt,alta,cr,d,lat,lata,lon,lona,mode,sbgts,und,used,vD,vDa,vE,vEa,vN,vNa,valid
	  53389 alt,cr,d,isdelay,pabs,pdif,sbgts,tas,temp,valid
	 109162 ax,ay,az,cr,d,dax,day,daz,dvx,dvy,dvz,gx,gy,gz,ok,rng,sbgts,temp,tst
	  54107 ax,ay,az,cr,d,mx,my,mz,ok,rng,sbgts,tst
	   5372 bl,cr,d,itow,p,pa,sbgts,status,th,tha,valid
	   1097 brk,buf_overrun,frame,overrun,parity,rx,tx
	   1120 class,id
	 133848 count,d,stamp,us
	   1096 count,wait,wait_max,work,work_max
	   5402 course,coursea,cr,d,itow,sbgts,status,type,vD,vDa,vE,vEa,vN,vNa
	   5482 cr,d,fixStat,flags,flags2,gpsFix,iTOW,msss,ttff
	 107810 cr,d,mode,p,pa,r,ra,sbgts,used,valid,y,ya
	   2199 d,devid,stamp,us
	1319424 d,stamp,us,value
	1330536 d,stamp,us,x,y,z
	      8 rest
	...
ch:
	 461017 4
edge_:
	   2192 1
	   2191 2
gpio_:
	   4383 18
gps:
	  19410 1
port:
	   7671 0
	1348186 2
	1483905 4
session:
	3677327 raspberrypi
srcid:
	   7671 0
	1483905 23
	1348186 92
start:
	      1 20240807111459.33007151
	3677327 20240807111459.330071510
trid:
	 444319 ACCEL_3G
	 443049 BAROP
	 443048 BAROT
	   2192 BMPOLL
	   2193 CANERR
	 216890 DTS
	 440542 GYRO_250DEG_S
	 452251 MAGN
	   6578 NAVSTATUS
	  69116 PULSERAISE
	   2193 RPI
	  54485 SBGAIR
	 108906 SBGEULER
	 110258 SBGIMU
	  55203 SBGMAG
	 107272 SBGNAV
	   1103 SBGREST
	   2193 TIOCGICOUNT
	 220178 TS
	...
Luuk-van-Dijks-MacBook-Pro:outflux lvd$ go build &&  gzcat ../dees/data/20240807111458/data.log.gz  | ./outflux
outflux:skipping record 3677328:tag at line 3677332:36: expected field key but none found
outflux:output 3677328 of 3677327 records.
MEASUREMENTS:
	3677328 debug
RECORD TYPES:
	  17587 V,d,stamp,us
	   1097 abrt,badmsg,bufsmall,crc,inv,noframe
	   5348 alt,alta,cr,d,diffAge,itow,lat,lata,lon,lona,numsv,sbgts,status,stid,type,und,used
	 106176 alt,alta,cr,d,lat,lata,lon,lona,mode,sbgts,und,used,vD,vDa,vE,vEa,vN,vNa,valid
	  53389 alt,cr,d,isdelay,pabs,pdif,sbgts,tas,temp,valid
	 109162 ax,ay,az,cr,d,dax,day,daz,dvx,dvy,dvz,gx,gy,gz,ok,rng,sbgts,temp,tst
	  54107 ax,ay,az,cr,d,mx,my,mz,ok,rng,sbgts,tst
	   5372 bl,cr,d,itow,p,pa,sbgts,status,th,tha,valid
	   1097 brk,buf_overrun,frame,overrun,parity,rx,tx
	   1120 class,id
	 133848 count,d,stamp,us
	   1096 count,wait,wait_max,work,work_max
	   5402 course,coursea,cr,d,itow,sbgts,status,type,vD,vDa,vE,vEa,vN,vNa
	   5482 cr,d,fixStat,flags,flags2,gpsFix,iTOW,msss,ttff
	 107810 cr,d,mode,p,pa,r,ra,sbgts,used,valid,y,ya
	   2199 d,devid,stamp,us
	1319424 d,stamp,us,value
	1330536 d,stamp,us,x,y,z
	      8 rest
	...
ch:
	 461017 4
edge_:
	   2192 1
	   2191 2
gpio_:
	   4383 18
gps:
	  19410 1
port:
	   7671 0
	1348186 2
	1483905 4
session:
	3677327 raspberrypi
srcid:
	   7671 0
	1483905 23
	1348186 92
start:
	      1 20240807111459.33007151
	3677327 20240807111459.330071510
trid:
	 444319 ACCEL_3G
	 443049 BAROP
	 443048 BAROT
	   2192 BMPOLL
	   2193 CANERR
	 216890 DTS
	 440542 GYRO_250DEG_S
	 452251 MAGN
	   6578 NAVSTATUS
	  69116 PULSERAISE
	   2193 RPI
	  54485 SBGAIR
	 108906 SBGEULER
	 110258 SBGIMU
	  55203 SBGMAG
	 107272 SBGNAV
	   1103 SBGREST
	   2193 TIOCGICOUNT
	 220178 TS
	...
```
