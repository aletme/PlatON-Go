package function.assembly;

import beforetest.ContractPrepareTest;
import network.platon.autotest.junit.annotations.DataSource;
import network.platon.autotest.junit.enums.DataSourceType;
import network.platon.contracts.AssemblyReturns;
import network.platon.utils.DataChangeUtil;
import org.junit.Before;
import org.junit.Test;
import org.web3j.protocol.core.methods.response.TransactionReceipt;
import org.web3j.tuples.generated.Tuple5;

import java.math.BigDecimal;
import java.math.BigInteger;

/**
 * @title 验证内联汇编关键字assembly,汇编赋值并返回多类型参数
 * @description:
 * @author: liweic
 * @create: 2020/01/07 19:01
 **/

public class AssemblyReturnsTest extends ContractPrepareTest {
    private String B;
    private String C;


    @Before
    public void before() {
        this.prepare();
        B = driverService.param.get("B");
        C = driverService.param.get("C");
    }

    @Test
    @DataSource(type = DataSourceType.EXCEL, file = "test.xls", sheetName = "Sheet1",
            author = "liweic", showName = "function.AssemblyReturnsTest-AssemblyReturns测试")
    public void Assemblyreturns() {
        try {
            AssemblyReturns assemblyreturns = AssemblyReturns.deploy(web3j, transactionManager, provider).send();

            String contractAddress = assemblyreturns.getContractAddress();
            TransactionReceipt tx = assemblyreturns.getTransactionReceipt().get();
            collector.logStepPass("AssemblyReturns deploy successfully.contractAddress:" + contractAddress + ", hash:" + tx.getTransactionHash());

            //验证AssemblyReturns
            Tuple5 result = assemblyreturns.f().send();

            collector.logStepPass("Assemblyreturns第一个返回值：" + result.getValue1());
            collector.assertEqual(new BigInteger("2") ,result.getValue1());

            byte[] value2 = (byte[])result.getValue2();
            String b = DataChangeUtil.bytesToHex(value2);
            collector.logStepPass("Assemblyreturns第二个返回值：" + b);
            collector.assertEqual(B ,b);

            byte[] value3 = (byte[])result.getValue3();
            String c = DataChangeUtil.bytesToHex(value3);
            collector.logStepPass("Assemblyreturns第三个返回值：" + c);
            collector.assertEqual(C ,c);

            collector.logStepPass("Assemblyreturns第四个返回值：" + result.getValue4());
            collector.assertEqual(true ,result.getValue4());

            collector.logStepPass("Assemblyreturns第五个返回值：" + result.getValue5());
            collector.assertEqual("0x1212121212121212121212121212121212121212" ,result.getValue5().toString());

        } catch (Exception e) {
            e.printStackTrace();
        }
    }
}

